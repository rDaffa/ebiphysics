// Copyright 2022 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/ebitengine/debugui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 640
	cellSizepx   = 20
)

type Game struct {
	canvasImage *ebiten.Image
	debugui     debugui.DebugUI
	ecs         Ecs
	cursor      Pos
	tick        int
	aa          bool
	showCenter  bool
}

type Entity struct {
	id int
}
type Component struct {
	name      *string
	visible   *bool
	position  *Pos
	mass      *float32
	inventory *[]Material
	balances  *[]Account
}
type velocity struct {
	speed     float32
	direction Pos
}
type Pos struct {
	x int
	y int
}
type Material struct {
	name string
	qty  int
}
type Account struct {
	name string
	qty  int
}
type Row struct {
	entity     Entity
	components Component
}

type Ecs struct {
	Table []Row
}

func (e *Ecs) addEntity(c Component) {
	// add a row to the ecs table
	EntityId := len(e.Table) + 1
	newEntity := Entity{EntityId}
	e.Table = append(e.Table, Row{newEntity, c})
}

func (e *Ecs) queryEntityName(name string) Component {
	var c Component
	for i := range e.Table {
		if name == *e.Table[i].components.name {
			c = e.Table[i].components
		} else {
			c = Component{}
		}
	}
	return c
}

func (g *Game) DrawOnGrid(screen *ebiten.Image) {
	for i := range g.ecs.Table {
		partPos := g.ecs.Table[i].components.position
		//		if g.tick%10 == 0 {
		max := 5
		min := -5
		ry := rand.New(rand.NewSource(time.Now().UnixNano()))
		partPos.y = partPos.y + ry.Intn(max-min+1) + min

		rx := rand.New(rand.NewSource(time.Now().UnixNano() + 2))
		partPos.x = partPos.x + rx.Intn(max-min+1) + min

		//		}

		vector.FillCircle(screen, float32(partPos.x), float32(partPos.y),
			cellSizepx, color.RGBA{0xff, 0, 0, 0xff}, g.aa)

		g.ecs.Table[i].components.position = partPos
	}
}

func (g *Game) Update() error {
	/*if _, err := g.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Lines", image.Rect(10, 10, 260, 160), func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
			ctx.Text(fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
			ctx.Checkbox(&g.aa, "Anti-aliasing")
			ctx.Checkbox(&g.showCenter, "Show center lines")
		})
		return nil
	}); err != nil {
		return err
	}*/
	mx, my := ebiten.CursorPosition()
	g.cursor = Pos{
		x: mx,
		y: my,
	}
	g.tick++

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.aa = !g.aa
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.showCenter = !g.showCenter
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.canvasImage, nil)
	msg := fmt.Sprintf("(%d, %d)", g.cursor.x, g.cursor.y)

	g.DrawOnGrid(screen)
	// g.ViewGrid(target)
	// g.debugui.Draw(screen)

	ebitenutil.DebugPrint(screen, msg)
	ebitenutil.DebugPrintAt(screen, msg, 0, 10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func initParticles() Ecs {
	e := Ecs{}
	for range 100 {
		visibility := true
		position := Pos{rand.Intn(screenWidth), rand.Intn(screenHeight)}
		e.addEntity(Component{
			visible:  &visibility,
			position: &position,
		})
	}
	return e
}

func NewGame() *Game {
	g := &Game{
		canvasImage: ebiten.NewImage(screenWidth, screenHeight),
		ecs:         initParticles(),
	}
	g.canvasImage.Fill(color.White)

	return g
}

func main() {
	g := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Lines (Ebitengine Demo)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
