/*
 *
 *           ____                        ______
 *          / __ \_________  _  ____  __/ ____/_  _____
 *         / /_/ / ___/ __ \| |/_/ / / / __/ / / / / _ \
 *        / ____/ /  / /_/ />  </ /_/ / /___/ /_/ /  __/
 *       /_/   /_/   \____/_/|_|\__, /_____/\__, /\___/
 *                                /_/         /_/
 *       ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
 *
 *                     Author: Suremeo (github.com/Suremeo)
 *
 *
 */

package console

import (
	"fmt"
	"github.com/Suremeo/ProxyEye/proxy/version"
	"github.com/fatih/color"
	"github.com/rs/zerolog/diode"
	"os"
	"strings"
	"time"
)

var dioder = diode.NewWriter(os.Stdout, 1000, 5*time.Millisecond, func(missed int) {
	fmt.Printf("Logger Dropped %d messages", missed)
})

var debug = true

var logo = fmt.Sprintf("\n\n          ____                        ______\n         / __ \\_________  _  ____  __/ ____/_  _____\n        / /_/ / ___/ __ \\| |/_/ / / / __/ / / / / _ \\\n       / ____/ /  / /_/ />  </ /_/ / /___/ /_/ /  __/\n      /_/   /_/   \\____/_/|_|\\__, /_____/\\__, /\\___/\n                               /_/         /_/\n      \n      %v\n \n                    %v %v\n\n", color.HiMagentaString("▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀"), color.HiWhiteString("Author:"), color.HiRedString("Vastle, LLC"))

func init() {
	_, _ = dioder.Write([]byte(fmt.Sprintf("\033]0;ProxyEye [%v]\007", version.String)))
	_, _ = dioder.Write([]byte(color.MagentaString(logo)))
}

func SetLevel(debugging bool) {
	debug = debugging
}

func Info(msg string, fields ...interface{}) {
	_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.HiMagentaString("INFO")+color.HiBlackString(" > "), fmt.Sprintf(msg, fields...), color.HiMagentaString)))
}

func Debug(msg string, fields ...interface{}) {
	if debug {
		_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.YellowString("DEBUG")+color.HiBlackString(" > "), fmt.Sprintf(msg, fields...), color.YellowString)))
	}
}

func Warn(msg string, fields ...interface{}) {
	_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.HiRedString("WARN")+color.HiBlackString(" > "), fmt.Sprintf(msg, fields...), color.HiRedString)))
}

func Error(msg string, err ...error) {
	if len(err) != 0 {
		if err[0] != nil {
			msg += " | Error: " + err[0].Error()
		}
	}
	_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.RedString("ERROR")+color.HiBlackString(" > "), msg, color.RedString)))
}

func Fatal(msg string, err ...error) {
	if len(err) != 0 {
		if err[0] != nil {
			msg += " | Error: " + err[0].Error()
		} else {
			return
		}
	}
	_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.RedString("FATAL")+color.HiBlackString(" > "), msg, color.RedString)))
	time.Sleep(100 * time.Millisecond)
	os.Exit(1)
}

func Panic(msg string, err ...error) {
	if len(err) != 0 {
		if err[0] != nil {
			msg += " | Error: " + err[0].Error()
		} else {
			return
		}
	}
	_, _ = dioder.Write([]byte(addPrefixToNewLine(getTime()+" "+color.RedString("PANIC")+color.HiBlackString(" > "), msg, color.RedString)))
	time.Sleep(100 * time.Millisecond)
	panic(msg)
}

// getTime returns a formatted verison of the current time.
func getTime() string {
	return color.HiBlackString(time.Now().Format("15:04:05"))
}

// addPrefixToNewLine adds the prefix provided to every single line of the provided body.
func addPrefixToNewLine(prefix, body string, textColor func(format string, a ...interface{}) string) string {
	var done string
	for _, line := range strings.Split(body, "\n") {
		done = done + "      " + prefix + textColor(line) + "\n"
	}
	return done
}
