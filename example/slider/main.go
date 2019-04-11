// using example/boxes as reference to figure out how to use package

package main

import (
	"log"
	"time"

	"github.com/dskinner/material"
	"github.com/dskinner/snd"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

var (
	env          = new(material.Environment)
	slider, btn2 *material.Button
	box1         *material.Material
	//	boxes [9]*material.Material
	sig   snd.Discrete
	quits []chan struct{}
)

func onStart(ctx gl.Context) {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})

	quits = []chan struct{}{}

	sig = make(snd.Discrete, len(material.ExpSig))
	copy(sig, material.ExpSig)
	rsig := make(snd.Discrete, len(material.ExpSig))
	copy(rsig, material.ExpSig)
	rsig.UnitInverse()
	sig = append(sig, rsig...)
	sig.NormalizeRange(0, 1)

	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)
	env.LoadGlyphs(ctx)

	slider = env.NewButton(ctx)
	slider.SetColor(material.Purple500)
	slider.OnTouch = setSlider

	btn2 = env.NewButton(ctx)
	btn2.SetColor(material.Teal500)
	btn2.OnTouch = resetSlider
	btn2.SetText("Reset")

	box1 = env.NewMaterial(ctx)
	box1.SetColor(material.LightBlue500)
	box1.Roundness = 5

}

// placeholder for function that will move the slider indicator up and down the
// slider
func setSlider(ev touch.Event) {
	m := box1.World()
	quits = append(quits, material.Animation{
		Sig:  sig,
		Dur:  1000 * time.Millisecond,
		Loop: false,
		Interp: func(dt float32) {
			m[1][3] = ev.Y * (1 - dt)
		},
		End: func() {
			m[1][3] = ev.Y
		},
	}.Do())
}

func resetSlider(ev touch.Event) {
	m := box1.World()
	y := m[1][3]
	m2 := slider.World()
	h2 := m2[1][1]
	y2 := m2[1][3]
	quits = append(quits, material.Animation{
		Sig:  sig,
		Dur:  500 * time.Millisecond,
		Loop: false,
		Start: func() {
			m[1][3] = y2
		},
		Interp: func(dt float32) {
			m[1][3] = y - y*0.6*dt
		},
		End: func() {
			m[1][3] = y2 + h2
		},
	}.Do())
}

//		quits = append(quits, material.Animation{
//			Sig:  sig,
//			Dur:  3000 * time.Millisecond,
//			Loop: false,
//			Interp: func(dt float32) {
//				m[2][3] = z + 4*dt
//			},
//		}.Do())
//	func() {
//		m := slider.World()
//		x, y := m[0][3], m[1][3]
//		w, h := m[0][0], m[1][1]
//		quits = append(quits, material.Animation{
//			Sig:  sig,
//			Dur:  4000 * time.Millisecond,
//			Loop: true,
//			Interp: func(dt float32) {
//				m[0][0] = w + 200*dt
//				m[0][3] = x - 200*dt/2
//				slider.SetText(fmt.Sprintf("w: %.2f\nh: %.2f", m[0][0], m[1][1]))
//			},
//		}.Do())
//		quits = append(quits, material.Animation{
//			Sig:  sig,
//			Dur:  2000 * time.Millisecond,
//			Loop: true,
//			Interp: func(dt float32) {
//				m[1][1] = h + 200*dt
//				m[1][3] = y - 200*dt/2
//			},
//		}.Do())
//	}()

func onStop(ctx gl.Context) {
	env.Unload(ctx)
	for _, q := range quits {
		q <- struct{}{}
	}
}

func onLayout(sz size.Event) {
	env.SetOrtho(sz)
	env.StartLayout()

	for _, q := range quits {
		q <- struct{}{}
	}
	quits = quits[:0]

	b, p := env.Box, env.Grid.Gutter
	env.AddConstraints(
		slider.Width(50), slider.Height(float32(sz.HeightPx)*0.6), slider.Z(1), slider.CenterVerticalIn(b), slider.EndIn(b, p),
		box1.Width(60), box1.Height(100), box1.Z(2), box1.CenterVerticalIn(b), box1.EndIn(b, p),
		btn2.Width(600), btn2.Height(200), btn2.Z(3), btn2.CenterHorizontalIn(b),
	)

	//is there a reason this was in onLayout and not onStart?
	//doesn't seem to affect text size predictably either way
	//slider.SetTextHeight(material.Dp(24).Px())
	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))
}

var lastpaint time.Time
var fps int

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey100.RGBA())
	ctx.Clear(gl.COLOR_BUFFER_BIT)
	env.Draw(ctx)
	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				if glctx == nil {
					a.Send(ev) // republish event until onStart is called
				} else {
					onLayout(ev)
				}
			case paint.Event:
				if glctx == nil || ev.External {
					continue
				}
				onPaint(glctx)
				a.Publish()
				a.Send(paint.Event{})
			case touch.Event:
				env.Touch(ev)
			}
		}
	})
}
