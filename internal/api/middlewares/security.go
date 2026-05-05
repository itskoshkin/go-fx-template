package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := ctx.Writer.Header()

		h.Set("Referrer-Policy", "strict-origin-when-cross-origin") // Send full referrer only on same-origin requests, and only the origin to HTTPS cross-origin targets
		h.Set("Content-Security-Policy", strings.Join([]string{     //MARK: Restricts which origins can load content
			"default-src 'self'",     // Default all fetchable resources to this origin
			"script-src 'self'",      // Allow only local scripts; Add inline scripts and trusted script CDN as "script-src 'self' 'unsafe-inline' https://cdn.example.com"
			"style-src 'self'",       // Allow local CSS; Add inline styles and trusted style CDN as "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com"
			"img-src 'self' data:",   // Allow local images and data URLs; Add trusted object storage/CDN origins as "img-src 'self' data: https://s3.example.com https://cdn.example.com"
			"font-src 'self'",        // Allow bundled fonts; Add font providers as "font-src 'self' https://fonts.example.com"
			"connect-src 'self'",     // Allow fetch/XHR/WebSocket only to this origin; Add API origins as "connect-src 'self' https://api.example.com wss://ws.example.com"
			"media-src 'self'",       // Allow local audio/video; Add trusted media CDN origins as "media-src 'self' https://media.example.com"
			"frame-src 'none'",       // Block iframes embedded by this app; Add trusted iframe origins as "frame-src 'self' https://www.youtube.com https://player.vimeo.com"
			"object-src 'none'",      // Block legacy plugin/embed/object content entirely; Add trusted plugin origins as "object-src 'self' https://plugins.example.com"
			"base-uri 'self'",        // Allow base URLs only from this origin; Add trusted base origins as "base-uri 'self' https://cdn.example.com"
			"form-action 'self'",     // Allow form submits only to this origin; Add payment/partner POST targets as "form-action 'self' https://checkout.example.com"
			"frame-ancestors 'none'", // Block embedding this app in iframes; Add trusted parent origins as "frame-ancestors 'self' https://partner.example.com"
		}, "; "))
		h.Set("X-Frame-Options", "DENY")                   // Legacy iframe protection for browsers that do not fully enforce CSP frame-ancestors
		h.Set("Permissions-Policy", strings.Join([]string{ //MARK: Disables listed browser capabilities; Replace () with (self) or an allowed origin (e.g. "geolocation=(\"https://maps.example.com\")") to enable a feature
			"accelerometer=()",                  // Disable device acceleration sensor access
			"ambient-light-sensor=()",           // Disable ambient light sensor access
			"aria-notify=()",                    // Disable experimental screen-reader announcement API access
			"attribution-reporting=()",          // Disable attribution/reporting APIs used for ad conversion measurement
			"autoplay=()",                       // Disable media autoplay unless explicitly re-enabled later
			"bluetooth=()",                      // Disable Web Bluetooth device access
			"browsing-topics=()",                // Disable Topics API interest-based advertising signals
			"camera=()",                         // Disable camera capture
			"captured-surface-control=()",       // Disable controls over captured display surfaces
			"ch-ua-high-entropy-values=()",      // Disable high-entropy User-Agent Client Hints
			"compute-pressure=()",               // Disable CPU/system pressure telemetry
			"cross-origin-isolated=()",          // Disable cross-origin isolation delegation for this document tree
			"deferred-fetch=()",                 // Disable deferred fetch quota for the top-level page
			"deferred-fetch-minimal=()",         // Disable minimal deferred fetch quota for subframes
			"display-capture=()",                // Disable screen/window/tab capture
			"encrypted-media=()",                // Disable Encrypted Media Extensions / DRM access
			"fullscreen=()",                     // Disable fullscreen requests
			"gamepad=()",                        // Disable Gamepad API access
			"geolocation=()",                    // Disable geolocation
			"gyroscope=()",                      // Disable device gyroscope sensor access
			"hid=()",                            // Disable WebHID device access
			"identity-credentials-get=()",       // Disable Federated Credential Management credential requests
			"idle-detection=()",                 // Disable user/device idle-state detection
			"language-detector=()",              // Disable browser-provided language detection APIs
			"local-fonts=()",                    // Disable enumeration of locally installed fonts
			"magnetometer=()",                   // Disable device magnetometer sensor access
			"microphone=()",                     // Disable microphone capture
			"midi=()",                           // Disable Web MIDI device access
			"on-device-speech-recognition=()",   // Disable local speech recognition APIs
			"otp-credentials=()",                // Disable WebOTP one-time-code credential access
			"payment=()",                        // Disable Payment Request API
			"picture-in-picture=()",             // Disable Picture-in-Picture video mode
			"private-state-token-issuance=()",   // Disable Private State Token issuance
			"private-state-token-redemption=()", // Disable Private State Token redemption
			"publickey-credentials-create=()",   // Disable passkey/WebAuthn credential creation
			"publickey-credentials-get=()",      // Disable passkey/WebAuthn credential retrieval
			"screen-wake-lock=()",               // Disable keeping the screen awake
			"serial=()",                         // Disable Web Serial device access
			"speaker-selection=()",              // Disable audio output device selection
			"storage-access=()",                 // Disable third-party iframe access to unpartitioned cookies/storage
			"summarizer=()",                     // Disable browser-provided summarization APIs
			"translator=()",                     // Disable browser-provided translation APIs
			"usb=()",                            // Disable WebUSB device access
			"web-share=()",                      // Disable native share sheet access
			"window-management=()",              // Disable multiscreen/window management APIs
			"xr-spatial-tracking=()",            // Disable WebXR spatial tracking
		}, ", "))
		h.Set("X-Content-Type-Options", "nosniff") // Prevent browsers from guessing MIME types and executing mislabeled content
		if ctx.Request.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=2592000; includeSubDomains") // Force browsers that reached us over HTTPS to use HTTPS for this host and subdomains for the next 30 days
		}

		ctx.Next()
	}
}
