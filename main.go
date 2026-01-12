package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Gotama struct {
	Nome            string    `json:"nome"`
	Nascimento      time.Time `json:"nascimento"`
	Fome            int       `json:"fome"`
	Carinho         int       `json:"carinho"`
	Elogios         int       `json:"elogios"`
	Felicidade      int       `json:"felicidade"`
	Estado          string    `json:"estado"`
	UltimaInteracao time.Time `json:"ultima_interacao"`
	CriadoEm        time.Time `json:"criado_em"`
}

func main() {
	g := &Gotama{
		Nome:       "Gintoki",
		Nascimento: time.Now(),
	}

	idleAnimation := []string{
		`
  .^._.^.
  | . . |
 (  ---  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
`,
		`
  .^._.^.
  | - - |
 (  ---  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
`,
	}

	runAnimation := []string{
		`
     .^._.^.
     | > < |
    (  ---  )
=3  .'     '.
=3  |/     \|
     \ /-\ /
      V   V
`,
		`
     .^._.^.
     | > < |
    (  ---  )
=3  .'     '.
=3  |/     \|
     \ /-\ /
      '   '
`,
	}

	app := tview.NewApplication()

	gameFrame := tview.NewTextView()

	gameView := tview.NewFlex().
		AddItem(gameFrame, 0, 1, false)

	gameView.SetBorder(true).SetTitle("Game")

	button := func(btn *tview.Button, height int) *tview.Flex {
		return tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(btn, height, 0, true).
			AddItem(nil, 0, 1, false)
	}

	info := tview.NewFlex().
		AddItem(tview.NewTextView().
			SetText(fmt.Sprintf("Nome: %s", g.Nome)), 0, 1, false).
		AddItem(tview.NewTextView().SetText("\n"), 0, 1, false).
		AddItem(tview.NewTextView().
			SetText(fmt.Sprintf("Idade: %v anos", g.Nascimento)), 0, 1, false)

	info.SetBorder(true).SetTitle("Info")

	resultText := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	result := tview.NewFlex().
		AddItem(resultText, 0, 1, false)

	result.SetBorder(true).SetTitle("Result")

	controls := tview.NewFlex().
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(
			button(tview.NewButton("Carinho").
				SetSelectedFunc(func() {
					g.Carinho += 1
					if g.Carinho <= 10 {
						resultText.SetText(fmt.Sprintf("%d Carinhos realizados", g.Carinho))
						gameFrame.Clear()
						gameFrame.SetText(`
  .^._.^.
  | ^ ^ |
 (  ヮ  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
						`)
					} else {
						resultText.SetText("Não é possivel realizar mais carinhos")
						gameFrame.Clear()
						gameFrame.SetText(`
  .^._.^.
  | > < |
 (  ###  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
						`)
					}
				}), 3), 20, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(
			button(tview.NewButton("Alimentar").
				SetSelectedFunc(func() {
					g.Fome += 1
					if g.Fome <= 5 {
						resultText.SetText(fmt.Sprintf("%d Alimentações Concluidas", g.Fome))
						gameFrame.Clear()
						gameFrame.SetText(`
  .^._.^.
  | ^ ^ |
 (  mmm  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
						`)
					} else {
						resultText.SetText("Gotama está cheio, não o alimente mais")
						gameFrame.Clear()
						gameFrame.SetText(`
  .^._.^.
  | > < |
 (  ###  )
 .'     '.
 |/     \|
  \ /-\ /
   V   V
						`)
					}
				}), 3), 0, 1, true).
		AddItem(tview.NewBox(), 2, 0, false)

	controls.SetBorder(true).SetTitle("Controls")

	ui := tview.NewGrid().
		SetRows(0, 10, 7).
		SetColumns(0, 35).
		SetBorders(false)

	ui.SetTitle("Gotama").SetBorder(true)

	ui.AddItem(gameView, 0, 0, 2, 1, 0, 0, false).
		AddItem(info, 0, 1, 1, 1, 0, 0, false).
		AddItem(result, 1, 1, 1, 1, 0, 0, false).
		AddItem(controls, 2, 0, 1, 2, 0, 0, true)

	var lastInteraction atomic.Int64

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		lastInteraction.Store(time.Now().UnixNano())
		return event
	})

	app.SetMouseCapture(func(
		event *tcell.EventMouse,
		action tview.MouseAction,
	) (*tcell.EventMouse, tview.MouseAction) {
		if event.Buttons() != tcell.ButtonNone {
			lastInteraction.Store(time.Now().UnixNano())
		}
		return event, action
	})

	const idleTimeout = 3 * time.Second

	go func() {
		for {
			var frames []string

			last := time.Unix(0, lastInteraction.Load())

			if time.Since(last) > idleTimeout {
				frames = idleAnimation
			} else {
				frames = runAnimation
			}

			for _, frame := range frames {
				app.QueueUpdateDraw(func() {
					gameFrame.SetText(frame)
				})
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	if err := app.SetRoot(ui, true).SetFocus(ui).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
