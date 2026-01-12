package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Gotama struct {
	Nome            string    `json:"nome"`
	Nascimento      time.Time `json:"nascimento"`
	Fase            string    `json:"fase"`
	Fome            int       `json:"fome"`
	Carinho         int       `json:"carinho"`
	Elogios         int       `json:"elogios"`
	Felicidade      int       `json:"felicidade"`
	Estado          string    `json:"estado"`
	UltimaInteracao time.Time `json:"ultima_interacao"`
	CriadoEm        time.Time `json:"criado_em"`
}

func LoadGotama() (*Gotama, error) {
	data, err := os.ReadFile("data/gotama.json")
	if err != nil {
		if os.IsNotExist(err) {
			now := time.Now()
			g := &Gotama{
				Nome:            "Gotama",
				Nascimento:      now,
				Fase:            "bebê",
				Fome:            50,
				Carinho:         50,
				Elogios:         0,
				Felicidade:      50,
				Estado:          "vivo",
				UltimaInteracao: now,
				CriadoEm:        now,
			}

			err = SaveGotama(g)
			return g, err
		}
		return nil, err
	}

	var g Gotama
	err = json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func SaveGotama(g *Gotama) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/gotama.json", data, 0644)
}

func UpdateState(g *Gotama) {
	calcularFase := func(idade time.Duration) string {
		dias := int(idade.Hours() / 24)

		switch {
		case dias < 7:
			return "bebê"
		case dias < 30:
			return "criança"
		case dias < 365:
			return "adulto"
		default:
			return "velho"
		}
	}
	now := time.Now()
	hours := int(now.Sub(g.UltimaInteracao).Hours())

	if hours > 0 {
		g.Fome += hours * 5
		g.Felicidade -= hours * 3

		if g.Fome > 100 {
			g.Fome = 100
			g.Estado = "morto"
		}
		if g.Felicidade < 0 {
			g.Felicidade = 0
		}

		g.UltimaInteracao = now
	}

	idade := now.Sub(g.Nascimento)
	g.Fase = calcularFase(idade)
}

func makeProgressBar(percent int, colorTag string) string {
	width := 20
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * float64(percent) / 100.0)
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	return fmt.Sprintf("[%s]%s[white] %3d%%", colorTag, bar, percent)
}

func main() {
	g, err := LoadGotama()
	if err != nil {
		panic(fmt.Sprintf("Erro ao carregar: %v", err))
	}

	app := tview.NewApplication()

	gameFrame := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetWrap(false)
	gameFrame.SetBorder(true).SetTitle(" Monitor ").SetTitleAlign(tview.AlignCenter)

	infoText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	textFome := tview.NewTextView().SetDynamicColors(true)
	textFelicidade := tview.NewTextView().SetDynamicColors(true)
	textCarinho := tview.NewTextView().SetDynamicColors(true)

	statsFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Fome:").SetTextAlign(tview.AlignLeft), 1, 0, false).
		AddItem(textFome, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(tview.NewTextView().SetText("Felicidade:").SetTextAlign(tview.AlignLeft), 1, 0, false).
		AddItem(textFelicidade, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(tview.NewTextView().SetText("Carinho:").SetTextAlign(tview.AlignLeft), 1, 0, false).
		AddItem(textCarinho, 1, 0, false)

	statsFlex.SetBorder(true).SetTitle(" Status Vitais ")

	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	logView.SetBorder(true).SetTitle(" Log ")

	logMessage := func(msg string) {
		_, err := fmt.Fprintf(logView, "[yellow]%s[white]: %s\n", time.Now().Format("15:04:05"), msg)
		if err != nil {
			panic(err)
		}
		logView.ScrollToEnd()
	}

	refreshUI := func() {
		idadeDias := int(time.Since(g.Nascimento).Hours() / 24)
		infoText.SetText(fmt.Sprintf("[::b]Nome:[::-] %s  |  [::b]Fase:[::-] %s  |  [::b]Idade:[::-] %d dias  |  [::b]Estado:[::-] %s",
			g.Nome, g.Fase, idadeDias, g.Estado))

		colorFome := "green"
		if g.Fome > 70 {
			colorFome = "red"
		} else if g.Fome > 40 {
			colorFome = "yellow"
		}
		textFome.SetText(makeProgressBar(g.Fome, colorFome))

		colorHappy := "green"
		if g.Felicidade < 30 {
			colorHappy = "red"
		}
		textFelicidade.SetText(makeProgressBar(g.Felicidade, colorHappy))

		textCarinho.SetText(makeProgressBar(g.Carinho, "blue"))
	}

	doAction := func(name string, action func(*Gotama)) {
		if g.Estado == "morto" {
			logMessage("O Gotama está morto. R.I.P.")
			return
		}
		action(g)
		logMessage(fmt.Sprintf("Ação: %s", name))
		refreshUI()
	}

	actAlimentar := func() {
		doAction("Alimentar", func(g *Gotama) {
			g.Fome -= 20
			if g.Fome < 0 {
				g.Fome = 0
			}
			g.Felicidade += 5
			g.UltimaInteracao = time.Now()
			err := SaveGotama(g)
			if err != nil {
				panic(err)
			}
		})
	}
	actElogiar := func() {
		doAction("Elogiar", func(g *Gotama) {
			g.Elogios++
			g.Carinho += 10
			g.Felicidade += 8
			g.UltimaInteracao = time.Now()
			err := SaveGotama(g)
			if err != nil {
				panic(err)
			}

		})
	}
	actCarinho := func() {
		doAction("Carinho", func(g *Gotama) {
			g.Carinho += 10
			g.Felicidade += 5
			g.Fome += 2
			if g.Carinho > 100 {
				g.Carinho = 100
			}
			if g.Felicidade > 100 {
				g.Felicidade = 100
			}
			err := SaveGotama(g)
			if err != nil {
				panic(err)
			}

		})
	}
	actBrincar := func() {
		doAction("Brincar", func(g *Gotama) {
			g.Felicidade += 15
			g.Fome += 10
			g.Carinho += 2
			if g.Felicidade > 100 {
				g.Felicidade = 100
			}
			if g.Carinho > 100 {
				g.Carinho = 100
			}
			err := SaveGotama(g)
			if err != nil {
				panic(err)
			}
		})
	}

	btnAlimentar := tview.NewButton("Alimentar(A)").SetSelectedFunc(actAlimentar)
	btnCarinho := tview.NewButton("Carinho(C)").SetSelectedFunc(actCarinho)
	btnBrincar := tview.NewButton("Brincar(B)").SetSelectedFunc(actBrincar)
	btnElogiar := tview.NewButton("Elogiar(E)").SetSelectedFunc(actElogiar)
	btnSair := tview.NewButton("Sair(Q)").SetSelectedFunc(func() { app.Stop() })

	for _, btn := range []*tview.Button{btnAlimentar, btnCarinho, btnBrincar, btnElogiar, btnSair} {
		btn.SetBackgroundColorActivated(tcell.ColorWhite)
		btn.SetLabelColorActivated(tcell.ColorBlack)
	}

	controls := tview.NewFlex().
		AddItem(btnAlimentar, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(btnCarinho, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(btnBrincar, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(btnElogiar, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(btnSair, 0, 1, false)

	controls.SetBorder(true).SetTitle(" Controles ")

	mainRow := tview.NewFlex().
		AddItem(gameFrame, 0, 2, false).
		AddItem(statsFlex, 30, 1, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(infoText, 1, 1, false).
		AddItem(mainRow, 0, 4, false).
		AddItem(logView, 6, 1, false).
		AddItem(controls, 3, 1, true)

	stagesFrame := map[string][]string{
		"ovo": {
			`
   .--.
  /    \
 |      |
  \    /
   '--'`,
			`
   .--.
  /  ? \
 | ?    |
  \    /
   '--'`,
		},
		"bebe": {
			`
   .^.
  ( o )
  / " \
 (     )
  '- -'`,
			`
   .^.
  ( - )
  / " \
 (     )
  '- -'`,
		},
		"criança": {
			`
 .^._.^.
 | . . |
(  ---  )
.'     '.
|/     \|
 \ /-\ /
  V   V`,
			`
 .^._.^.
 | ^ ^ |
(  ~~~  )
.'     '.
|/     \|
 \ /-\ /
  V   V`,
		},
		"adulto": {
			`
  .^._.^.
  | o o |
 (  ---  )
 .'  |  '.
 |/  |  \|
 |   |   |
  \ /-\ /
   V   V`,
			`
  .^._.^.
  | - - |
 (  ---  )
 .'  |  '.
 |/  |  \|
 |   |   |
  \ /-\ /
   V   V`,
		},
		"doente": {
			`
 .^._.^.
 | x x |
(   o   )===|
.'     '.
|/     \|
 \ /-\ /
  V   V`,
			`
 .^._.^.
 | @ @ |
(  ~~~  )===|
.'     '.
|/     \|
 \ /-\ /
  V   V`,
		},
		"morto": {
			`
  .----.
 /      \
|  R.I.P |
|    +   |
|        |
_|________|_`,
			`
   _ _
  (o o)
  | ~ |
  '---'
  .----.
_|________|_`,
		},
		"bebe_doente": {
			`
   .^.
  ( > )
  / ~ \
 (     )
  '- -'`,
			`
   .^.
  ( x )
  / ~ \
 (     )
  '- -'`},
		"bebe_morto": {
			`
     +
   .-.-.
  | RIP |
  '-----'`,
			`
     +
   .-.-.
  | RIP |
  '-----'`},
		"crianca_doente": {
			`
 .^._.^.
 | - - |
(   ~   ) --|
.'     '.`,
			`
 .^._.^.
 | @ @ |
(   o   ) --|
.'     '.`},
		"crianca_morta": {
			`
  .----.
 /      \
|  R.I.P |
| kiddo  |`,
			`
     @
  .----|
 /     | \
|  R.I.P |
| kiddo  |`},
	}

	const canvasWidth = 110

	go func() {
		for {
			keyFase := g.Fase
			if g.Fase == "bebe" || g.Fase == "criança" || g.Fase == "adulto" {
				if g.Estado == "morto" {
					keyFase += "_morto"
				} else if g.Estado == "doente" || (g.Fome > 80) {
					keyFase += "_doente"
				}
			}

			frames, ok := stagesFrame[keyFase]
			if !ok {
				frames, ok = stagesFrame[g.Fase]
				if !ok {
					frames = stagesFrame["ovo"]
				}
			}

			for _, frame := range frames {
				maxLineWidth := func(s string) int {
					max := 0
					for _, l := range strings.Split(strings.Trim(s, "\n"), "\n") {
						if len(l) > max {
							max = len(l)
						}
					}
					return max
				}

				centerASCII := func(frame string, canvasWidth int) string {
					lines := strings.Split(strings.Trim(frame, "\n"), "\n")
					spriteWidth := maxLineWidth(frame)

					offset := (canvasWidth - spriteWidth) / 2
					if offset < 0 {
						offset = 0
					}

					for i, l := range lines {
						lines[i] = strings.Repeat(" ", offset) + l
					}

					return strings.Join(lines, "\n")
				}

				app.QueueUpdateDraw(func() {
					centered := centerASCII(frame, canvasWidth)
					gameFrame.SetText("\n\n" + centered)
					refreshUI()
				})
				time.Sleep(800 * time.Millisecond)
			}
		}
	}()

	go func() {
		UpdateState(g)
		app.QueueUpdateDraw(func() { refreshUI() })
	}()

	var lastInteraction atomic.Int64
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		lastInteraction.Store(time.Now().UnixNano())
		switch event.Rune() {
		case 'a', 'A':
			actAlimentar()
		case 'c', 'C':
			actCarinho()
		case 'b', 'B':
			actBrincar()
		case 'e', 'E':
			actElogiar()
		case 'q', 'Q':
			app.Stop()
		}
		return event
	})

	app.SetMouseCapture(func(ev *tcell.EventMouse, ac tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if ev.Buttons() != tcell.ButtonNone {
			lastInteraction.Store(time.Now().UnixNano())
		}
		return ev, ac
	})

	refreshUI()
	app.SetFocus(controls)

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
