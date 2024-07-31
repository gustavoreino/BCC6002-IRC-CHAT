package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/net/websocket"
)

type C = layout.Context
type D = layout.Dimensions

func main() {
	serverURL := "ws://localhost:8000/ws" // Replace with your WebSocket server URL
	ws, errWs := websocket.Dial(serverURL, "", "ws://localhost:3001")
	if errWs != nil {
		log.Fatal("Error connecting to WebSocket server:", errWs)
	}

	// Send a message to the WebSocket server
	// msg := "Hello, WebSocket server!"
	// _, errWs = ws.Write([]byte(msg))
	// if errWs != nil {
	// 	log.Println("Error sending message:", errWs)
	// 	return
	// }

	// Keep the main goroutine alive
	var response = make([]byte, 1024)
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Title("IrcClient"))
		w.Option(app.Size(unit.Dp(400), unit.Dp(600)))
		if err := draw(w, ws, response); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
	defer ws.Close()
}

func draw(w *app.Window, ws *websocket.Conn, response []byte) error {
	// ...
	var ops op.Ops
	var sendButton widget.Clickable
	var textInput widget.Editor
	var msgs []string

	// th defines the material design style
	th := material.NewTheme()

	var pastText string

	// listen for events in the window.
	for {

		// then detect the type
		switch typ := w.Event().(type) {

		// this is sent when the application should re-render.
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)
			// if sendButton.Clicked(gtx) {
			// 	fmt.Println(strings.TrimSpace(textInput.Text()))
			// 	sendMessage(ws, strings.TrimSpace(textInput.Text()))
			// }
			go func(ws *websocket.Conn) {
				n, err := ws.Read(response)
				if err != nil {
					log.Println("Error reading message:", err)
				}
				fmt.Printf("Received message: %s\n", response[:n])
				var newMsg = string(response)
				msgs = append(msgs, newMsg)
				fmt.Println(newMsg)
			}(ws)

			if pastText != strings.TrimSpace(textInput.Text()) {
				pastText = strings.TrimSpace(textInput.Text())
				sendMessage(ws, pastText)
			}

			// Let's try out the flexbox layout:
			layout.Flex{
				// Vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						// Wrap the editor in material design
						ed := material.Editor(th, &textInput, "Texto")
						// textInput.SingleLine = false
						textInput.MaxLen = 96
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						border := widget.Border{
							Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						// ... before laying it out, one inside the other
						return margins.Layout(gtx,
							func(gtx C) D {
								gtx.Constraints.Max.Y = int(60)
								gtx.Constraints.Min.Y = int(60)
								gtx.Constraints.Max.X = 330
								gtx.Constraints.Min.X = 330
								return border.Layout(gtx, ed.Layout)
							},
						)
					},
				),
				// First one to hold a button ...
				layout.Rigid(
					func(gtx C) D {
						// ONE: First define margins around the button using layout.Inset ...
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						// TWO: ... then we lay out those margins ...
						return margins.Layout(gtx,
							// THREE: ... and finally within the margins, we ddefine and lay out the button
							func(gtx C) D {
								btn := material.Button(th, &sendButton, "Enviar")
								return btn.Layout(gtx)
							},
						)
					},
				),
			)
			typ.Frame(gtx.Ops)
		// this is sent when the application is closed
		case app.DestroyEvent:
			return typ.Err
		}
	}
}

func sendMessage(ws *websocket.Conn, msg string) {
	var err error
	_, err = ws.Write([]byte(msg))
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}
}
