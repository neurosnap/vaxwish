package vaxwish

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

func VaxisMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sesh ssh.Session) {
			pty, windowChanges, ok := sesh.Pty()
			if !ok {
				next(sesh)
				return
			}

			vx, err := vaxis.New(vaxis.Options{
				WithTTYFile: pty.Master,
			})
			if err != nil {
				wish.Errorln(sesh, err)
				next(sesh)
				return
			}
			defer vx.Close()

			win := vx.Window()
			win.Width = pty.Window.Width
			win.Height = pty.Window.Height

			go func() {
				for {
					select {
					case w := <-windowChanges:
						win.Width = w.Width
						win.Height = w.Height
						vx.Resize()
					}
				}
			}()

			for ev := range vx.Events() {
				switch ev := ev.(type) {
				case vaxis.Key:
					switch ev.String() {
					case "Ctrl+c":
						next(sesh)
						return
					}
				}

				win := vx.Window()
				win.Width = pty.Window.Width
				win.Height = pty.Window.Height
				win.Clear()
				win.Print(vaxis.Segment{Text: "Hello, World!"})
				vx.Render()
			}
		}
	}
}
