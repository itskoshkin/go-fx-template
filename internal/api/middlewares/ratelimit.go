package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const parsedJSONBodyKey = "parsed_json_body"

var (
	LoginRateLimitPolicy = RateLimitPolicy{
		name: "login",
		ipBucket: &tokenBucketPolicy{
			ratePerSecond: 1,
			burst:         5,
		},
		subjectWindow: &fixedWindowPolicy{
			field:  "username",
			limit:  20,
			window: 10 * time.Minute,
		},
		failureBackoff: &failureBackoffPolicy{
			field:           "username",
			window:          10 * time.Minute,
			firstThreshold:  5,
			firstBlock:      5 * time.Minute,
			secondThreshold: 10,
			secondBlock:     30 * time.Minute,
		},
	}
	RefreshRateLimitPolicy = RateLimitPolicy{
		name: "refresh",
		ipBucket: &tokenBucketPolicy{
			ratePerSecond: 2,
			burst:         10,
		},
	}
	ForgotPasswordRateLimitPolicy = RateLimitPolicy{
		name: "forgot-password",
		ipBucket: &tokenBucketPolicy{
			ratePerSecond: 1.0 / 30.0,
			burst:         3,
		},
		subjectWindow: &fixedWindowPolicy{
			field:  "email",
			limit:  3,
			window: time.Hour,
		},
	}
)

type RateLimitPolicy struct {
	name           string
	ipBucket       *tokenBucketPolicy
	subjectWindow  *fixedWindowPolicy
	failureBackoff *failureBackoffPolicy
}

type tokenBucketPolicy struct {
	ratePerSecond float64
	burst         int
}

type fixedWindowPolicy struct {
	field  string
	limit  int
	window time.Duration
}

type failureBackoffPolicy struct {
	field           string
	window          time.Duration
	firstThreshold  int
	firstBlock      time.Duration
	secondThreshold int
	secondBlock     time.Duration
}

type tokenBucket struct {
	tokens    float64
	updatedAt time.Time
	lastSeen  time.Time
}

type fixedWindowCounter struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

type failureBackoffState struct {
	failures      int
	windowResetAt time.Time
	blockedUntil  time.Time
	lastSeen      time.Time
}

type authRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*tokenBucket
	windows     map[string]*fixedWindowCounter
	backoffs    map[string]*failureBackoffState
	lastCleanup time.Time
}

func newAuthRateLimiter() *authRateLimiter {
	return &authRateLimiter{
		buckets:     make(map[string]*tokenBucket),
		windows:     make(map[string]*fixedWindowCounter),
		backoffs:    make(map[string]*failureBackoffState),
		lastCleanup: time.Now(),
	}
}

func (mw *Middlewares) RateLimit(policy RateLimitPolicy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		now := time.Now()
		ip := clientIPForRateLimit(ctx)

		if policy.ipBucket != nil && ip != "" {
			allowed, retryAfter := mw.limiter.allowTokenBucket(policy.name+":ip:"+ip, policy.ipBucket.ratePerSecond, policy.ipBucket.burst, now)
			if !allowed {
				abortWithTooManyRequests(ctx, retryAfter)
				return
			}
		}

		subject := ""
		if policy.subjectWindow != nil {
			subject = normalizedBodyField(ctx, policy.subjectWindow.field)
			if subject != "" {
				allowed, retryAfter := mw.limiter.allowFixedWindow(policy.name+":"+policy.subjectWindow.field+":"+subject, policy.subjectWindow.limit, policy.subjectWindow.window, now)
				if !allowed {
					abortWithTooManyRequests(ctx, retryAfter)
					return
				}
			}
		}

		backoffSubject := ""
		if policy.failureBackoff != nil {
			if policy.subjectWindow != nil && policy.subjectWindow.field == policy.failureBackoff.field {
				backoffSubject = subject
			} else {
				backoffSubject = normalizedBodyField(ctx, policy.failureBackoff.field)
			}

			if backoffSubject != "" {
				retryAfter := mw.limiter.failureBlockRetry(backoffKey(policy.name, backoffSubject, ip), now)
				if retryAfter > 0 {
					abortWithTooManyRequests(ctx, retryAfter)
					return
				}
			}
		}

		ctx.Next()

		if policy.failureBackoff == nil || backoffSubject == "" {
			return
		}

		key := backoffKey(policy.name, backoffSubject, ip)
		switch ctx.Writer.Status() {
		case http.StatusOK:
			mw.limiter.clearFailureBackoff(key)
		case http.StatusUnauthorized:
			mw.limiter.recordFailureBackoff(
				key,
				time.Now(),
				policy.failureBackoff.window,
				policy.failureBackoff.firstThreshold,
				policy.failureBackoff.secondThreshold,
				policy.failureBackoff.firstBlock,
				policy.failureBackoff.secondBlock,
			)
		}
	}
}

func (rl *authRateLimiter) allowTokenBucket(key string, ratePerSecond float64, burst int, now time.Time) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanup(now)

	bucket, ok := rl.buckets[key]
	if !ok {
		bucket = &tokenBucket{
			tokens:    float64(burst),
			updatedAt: now,
			lastSeen:  now,
		}
		rl.buckets[key] = bucket
	}

	elapsed := now.Sub(bucket.updatedAt).Seconds()
	bucket.tokens = math.Min(float64(burst), bucket.tokens+elapsed*ratePerSecond)
	bucket.updatedAt = now
	bucket.lastSeen = now

	if bucket.tokens < 1 {
		retryAfter := int(math.Ceil((1 - bucket.tokens) / ratePerSecond))
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}

	bucket.tokens--
	return true, 0
}

func (rl *authRateLimiter) allowFixedWindow(key string, limit int, window time.Duration, now time.Time) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanup(now)

	counter, ok := rl.windows[key]
	if !ok || !now.Before(counter.resetAt) {
		counter = &fixedWindowCounter{
			count:    0,
			resetAt:  now.Add(window),
			lastSeen: now,
		}
		rl.windows[key] = counter
	}

	counter.lastSeen = now
	if counter.count >= limit {
		retryAfter := int(math.Ceil(counter.resetAt.Sub(now).Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}

	counter.count++
	return true, 0
}

func (rl *authRateLimiter) failureBlockRetry(key string, now time.Time) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanup(now)

	state, ok := rl.backoffs[key]
	if !ok || !now.Before(state.blockedUntil) {
		return 0
	}

	retryAfter := int(math.Ceil(state.blockedUntil.Sub(now).Seconds()))
	if retryAfter < 1 {
		retryAfter = 1
	}

	state.lastSeen = now
	return retryAfter
}

func (rl *authRateLimiter) recordFailureBackoff(key string, now time.Time, window time.Duration, firstThreshold, secondThreshold int, firstBlock, secondBlock time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanup(now)

	state, ok := rl.backoffs[key]
	if !ok || !now.Before(state.windowResetAt) {
		state = &failureBackoffState{
			failures:      0,
			windowResetAt: now.Add(window),
			lastSeen:      now,
		}
		rl.backoffs[key] = state
	}

	state.failures++
	state.lastSeen = now

	switch {
	case state.failures >= secondThreshold:
		if now.Add(secondBlock).After(state.blockedUntil) {
			state.blockedUntil = now.Add(secondBlock)
		}
	case state.failures >= firstThreshold:
		if now.Add(firstBlock).After(state.blockedUntil) {
			state.blockedUntil = now.Add(firstBlock)
		}
	}
}

func (rl *authRateLimiter) clearFailureBackoff(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.backoffs, key)
}

func (rl *authRateLimiter) cleanup(now time.Time) {
	if now.Sub(rl.lastCleanup) < 5*time.Minute {
		return
	}

	for key, bucket := range rl.buckets {
		if now.Sub(bucket.lastSeen) > 15*time.Minute {
			delete(rl.buckets, key)
		}
	}

	for key, counter := range rl.windows {
		if now.Sub(counter.lastSeen) > 2*time.Hour {
			delete(rl.windows, key)
		}
	}

	for key, state := range rl.backoffs {
		if now.Sub(state.lastSeen) > time.Hour && !now.Before(state.blockedUntil) {
			delete(rl.backoffs, key)
		}
	}

	rl.lastCleanup = now
}

func abortWithTooManyRequests(ctx *gin.Context, retryAfter int) {
	if retryAfter > 0 {
		ctx.Header("Retry-After", strconv.Itoa(retryAfter))
	}

	ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "too many requests"})
}

func clientIPForRateLimit(ctx *gin.Context) string {
	ip := strings.TrimSpace(ctx.ClientIP())
	if ip != "" {
		return ip
	}

	return strings.TrimSpace(ctx.Request.RemoteAddr)
}

func backoffKey(scope, subject, ip string) string {
	if ip == "" {
		return scope + ":" + subject
	}

	return scope + ":" + subject + ":" + ip
}

func normalizedBodyField(ctx *gin.Context, field string) string {
	body, ok := cachedJSONBody(ctx)
	if !ok {
		return ""
	}

	value, ok := body[field].(string)
	if !ok {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(value))
}

func cachedJSONBody(ctx *gin.Context) (map[string]any, bool) {
	if value, ok := ctx.Get(parsedJSONBodyKey); ok {
		body, ok := value.(map[string]any)
		return body, ok
	}

	if ctx.Request.Body == nil {
		return nil, false
	}

	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return nil, false
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(rawBody))

	var body map[string]any
	if err = json.Unmarshal(rawBody, &body); err != nil {
		return nil, false
	}

	ctx.Set(parsedJSONBodyKey, body)
	return body, true
}
