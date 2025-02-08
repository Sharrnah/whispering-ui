package Logging

import (
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"runtime/debug"
	"time"
)

const Dsn = "https://2978cf7a0b7c45c0abd5b1249d15a157@glitchtip.libs.space/1"
const FlushTimeoutDefault = 3 * time.Second

var ReportingEnabled = false

func EnableReporting(enable bool) {
	ReportingEnabled = enable
	fyne.CurrentApp().Preferences().SetBool("SendErrorsToServer", ReportingEnabled)
}

func IsReportingEnabled() bool {
	return ReportingEnabled
}

func PanicLogger() {
	if r := recover(); r != nil {
		// 2. Create a log file when a crash occurs
		logFile, err := os.OpenFile("error_ui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Failed to create log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()

		// Use logFile as the output for the log package
		log.SetOutput(logFile)

		// 3. Capture the stack trace and log the error details
		stackTrace := debug.Stack()
		log.Printf("Panic occurred: %v\nStack trace:\n%s", r, stackTrace)

		fmt.Println("A crash occurred. Check the error_ui.log file for more information.")
	}
}

func ErrorHandlerInit(version string) {
	//if ReportingEnabled {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              Dsn,
		Release:          version,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Printf("sentry.Init: %s", err)
	}
	//} else {
	//	err := sentry.Init(sentry.ClientOptions{
	//		Dsn: "",
	//	})
	//	if err != nil {
	//		log.Printf("sentry.Init: %s", err)
	//	}
	//}
}

func GoRoutineErrorHandler(scopeConfig func(scope *sentry.Scope)) {
	defer ErrorHandlerRecover()
	localHub := CloneHub()
	localHub.WithScope(scopeConfig)
	//ConfigureScope(localHub, scopeConfig)
}

func CurrentHub() *sentry.Hub {
	return sentry.CurrentHub()
}

// CloneHub creates a new Hub that shares the same client with the current Hub.
// needed in Goroutines. Make sure to check if Logging is enabled when working in a cloned hub (WithScope) especially when calling Flush on the hub. [localHub.Flush()]
func CloneHub() *sentry.Hub {
	return CurrentHub().Clone()
}

func ConfigureScope(hub *sentry.Hub, scopeConfig func(scope *sentry.Scope)) {
	hub.ConfigureScope(scopeConfig)
}

// ErrorHandlerRecover recovers from a panic and flushes any buffered events to Sentry.
// It ensures that any errors captured by Sentry are sent before the application exits.
// This function does not take any parameters and does not return any values.
func ErrorHandlerRecover() {
	defer PanicLogger()

	if ReportingEnabled {
		sentry.Recover()
		Flush(FlushTimeoutDefault)
	}
}

func Flush(timeout time.Duration) {
	if ReportingEnabled {
		sentry.Flush(timeout)
	}
}

// AddBreadcrumb records a new breadcrumb.
//
// The total number of breadcrumbs that can be recorded are limited by the
// configuration on the client.
func AddBreadcrumb(breadcrumb *sentry.Breadcrumb) {
	sentry.AddBreadcrumb(breadcrumb)
}

// CaptureMessage captures an arbitrary message.
func CaptureMessage(message string) {
	if ReportingEnabled {
		sentry.CaptureMessage(message)
	}
}

// CaptureException captures an error.
func CaptureException(exception error) {
	if ReportingEnabled {
		sentry.CaptureException(exception)
	}
}
