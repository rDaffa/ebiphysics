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
	name            *string
	visible         *bool
	position        *Pos
	mass            *float32
	velocity        *Velocity
	collisionRadius *int
	borderLine      *BorderLine
}
type BorderLine struct {
	start Pos
	end   Pos
}
type Velocity struct {
	speed     float32
	direction Pos
}
type Pos struct {
	x int
	y int
}

type EntityRow struct {
	entity     Entity
	components Component
}

type Ecs struct {
	Entities []EntityRow
}

type system struct {
	particleMotion *ParticleMotion
}
type ParticleMotion struct{}

func (e *Ecs) addEntity(c Component) error {
	// add a row to the ecs table
	EntityId := len(e.Entities) + 1
	newEntity := Entity{EntityId}
	e.Entities = append(e.Entities, EntityRow{newEntity, c})
	return nil
}

func (e *Ecs) queryEntityName(name string) Component {
	var c Component
	for i := range e.Entities {
		if name == *e.Entities[i].components.name {
			c = e.Entities[i].components
		} else {
			c = Component{}
		}
	}
	return c
}

// to do make filter function for entities

func randNum(min, max int, salt int64) int {
	rnd := rand.New(rand.NewSource(
		time.Now().UnixNano() + salt))
	return rnd.Intn(max-min+1) + min
}

/*
	func checkCollision(ents []EntityRow) {
		for i := range ents {
			if ents[i].components.mass != nil &&
				ents[i].components.velocity != nil &&
				ents[i].components.collisionRadius != nil {
			}
		}
	}
*/
func (g *Game) drawLoop(screen *ebiten.Image) {
	ents := g.ecs.Entities

	for i := range ents {
		moveEntities(ents[i].components)
		drawEntities(screen, ents[i].components)
	}
}

func moveEntities(c Component) {
	if c.visible != nil &&
		c.velocity != nil &&
		c.name != nil {
		if *c.name == "particle" {
			c.position.x += c.velocity.direction.x * int(c.velocity.speed)
			c.position.y += c.velocity.direction.y * int(c.velocity.speed)
		}
	}
}

func drawEntities(s *ebiten.Image, c Component) {
	if c.visible != nil &&
		c.velocity != nil &&
		c.name != nil {
		if *c.name == "particle" {
			vector.FillCircle(s, float32(c.position.x), float32(c.position.y),
				cellSizepx, color.RGBA{0xff, 0, 0, 0xff}, true)
		}
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
	g.drawLoop(screen)
	// g.debugui.Draw(screen)

	ebitenutil.DebugPrint(screen, msg)
	ebitenutil.DebugPrintAt(screen, msg, 0, 10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func initParticles() Ecs {
	e := Ecs{}
	for i := range 100 {
		visibility := true
		position := Pos{rand.Intn(screenWidth), rand.Intn(screenHeight)}
		velocity := Velocity{2, Pos{randNum(-1, 1, int64(1+i)), randNum(-1, 1, int64(3*i))}}
		name := "particle"
		err := e.addEntity(Component{
			name:     &name,
			visible:  &visibility,
			position: &position,
			velocity: &velocity,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	name := "left wall"
	line := BorderLine{
		Pos{0, 0},
		Pos{0, screenHeight},
	}
	err := e.addEntity(Component{
		name:       &name,
		borderLine: &line,
	})
	if err != nil {
		log.Fatal(err)
	}

	name = "right wall"
	line = BorderLine{
		Pos{screenWidth, 0},
		Pos{screenWidth, screenHeight},
	}
	e.addEntity(Component{
		name:       &name,
		borderLine: &line,
	})
	name = "top wall"
	line = BorderLine{
		Pos{0, 0},
		Pos{screenWidth, 0},
	}
	e.addEntity(Component{
		name:       &name,
		borderLine: &line,
	})
	name = "bottom wall"
	line = BorderLine{
		Pos{0, screenHeight},
		Pos{screenWidth, screenHeight},
	}
	e.addEntity(Component{
		name:       &name,
		borderLine: &line,
	})

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
