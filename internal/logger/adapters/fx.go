package adapters

import (
	"strings"

	"go.uber.org/fx/fxevent"

	"go-fx-template/internal/logger"
)

type CustomFxLogger struct{}

func (CustomFxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		logger.Debug("Fx: %s starting", shortFn(e.FunctionName))
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			logger.Error("Fx: %s OnStart failed: %v", shortFn(e.FunctionName), e.Err)
		} else {
			logger.Debug("Fx: %s started in %s", shortFn(e.FunctionName), e.Runtime)
		}
	case *fxevent.OnStopExecuting:
		logger.Debug("Fx: %s stopping", shortFn(e.FunctionName))
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			logger.Error("Fx: %s OnStop failed: %v", shortFn(e.FunctionName), e.Err)
		} else {
			logger.Debug("Fx: %s stopped in %s", shortFn(e.FunctionName), e.Runtime)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			logger.Error("Fx: supply %s failed: %v", e.TypeName, e.Err)
		} else {
			logger.Debug("Fx: supplied %s", e.TypeName)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			logger.Error("Fx: provide %s failed: %v", shortFn(e.ConstructorName), e.Err)
			return
		}
		for _, t := range e.OutputTypeNames {
			logger.Debug("Fx: %s -> %s", shortFn(e.ConstructorName), t)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			logger.Error("Fx: decorator %s failed: %v", shortFn(e.DecoratorName), e.Err)
		}
	case *fxevent.Replaced:
		if e.Err != nil {
			logger.Error("Fx: replace failed: %v", e.Err)
		}
	case *fxevent.Invoking:
		logger.Debug("Fx: invoking %s", shortFn(e.FunctionName))
	case *fxevent.Invoked:
		if e.Err != nil {
			logger.Error("Fx: invoke %s failed: %v", shortFn(e.FunctionName), e.Err)
		}
	case *fxevent.Stopping:
		logger.Info("Fx: received %s signal, stopping", strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			logger.Error("Fx: stop failed: %v", e.Err)
		}
	case *fxevent.RollingBack:
		logger.Error("Fx: rolling back after start failure: %v", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			logger.Error("Fx: rollback failed: %v", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			logger.Error("Fx: application start failed: %v", e.Err)
		} else {
			logger.Info("Fx: application started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			logger.Error("Fx: logger init failed: %v", e.Err)
		}
	}
}

func shortFn(name string) string {
	//name = strings.TrimSuffix(name, "()")
	//if i := strings.LastIndex(name, "/"); i >= 0 {
	//	name = name[i+1:]
	//}
	return name
}
