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
	cellSizepx   = 10
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
	color           *color.RGBA
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
	res := rnd.Intn(max-min+1) + min
	if res == 0 {
		res = 1
	}
	return res
}

func (g *Game) drawLoop(screen *ebiten.Image) {
	ents := g.ecs.Entities
	for i := range ents {
		moveEntities(ents[i].entity.id, ents[i].components, ents)
		drawEntities(screen, ents[i].components)
	}
}

// SYSTEMS
func moveEntities(id int, c Component, p []EntityRow) {
	if c.visible != nil &&
		c.velocity != nil &&
		c.name != nil {
		if *c.name == "particle" {
			particleCollision(id, c, p)

			boundaryCollision(c)

			particleNewPos(c)
		}
	}
}

func particleNewPos(c Component) {
	c.position.x += c.velocity.direction.x * int(c.velocity.speed)
	c.position.y += c.velocity.direction.y * int(c.velocity.speed)
}

func boundaryCollision(c Component) {
	radius := cellSizepx / 2
	if c.position.x-radius < 0 || c.position.x+radius > screenWidth {
		c.velocity.direction.x = -c.velocity.direction.x
	}
	if c.position.y-radius < 0 || c.position.y+radius > screenHeight {
		c.velocity.direction.y = -c.velocity.direction.y
	}
}

func particleCollision(id int, c Component, p []EntityRow) {
	radius := cellSizepx / 2
	for i := range p {
		if *p[i].components.name == "particle" && p[i].entity.id != id {
			Xc := c.position.x
			Yc := c.position.y
			Xp := p[i].components.position.x
			Yp := p[i].components.position.y

			if Xc-Xp+2*radius > 0 && Xc-Xp-2*radius < 0 && Yc-Yp+2*radius > 0 && Yc-Yp-2*radius < 0 {
				//EXCHANGING VELOCITIES DOESNT WORK YET, NEED TO DO PAIR LOOP NOT ONE BY ONE
				//completely inelastic colisions
				// should add different masses
				// should add different velocities and colours
				/*
					c.velocity.direction.x = p[i].components.velocity.direction.x
					c.velocity.direction.y = p[i].components.velocity.direction.y
				*/
				c.velocity.direction.x = -c.velocity.direction.x
				c.velocity.direction.y = -c.velocity.direction.y
			}
		}
	}
}

func drawline(s *ebiten.Image, b BorderLine) {
	var path vector.Path
	path.MoveTo(float32(b.start.x), float32(b.start.y))
	path.LineTo(float32(b.end.x), float32(b.end.y))
	strokeOp := &vector.StrokeOptions{}
	strokeOp.LineCap = 1
	strokeOp.LineJoin = 1
	strokeOp.MiterLimit = 1
	strokeOp.Width = float32(3)
	drawOp := &vector.DrawPathOptions{}
	drawOp.AntiAlias = true
	drawOp.ColorScale.ScaleWithColor(color.RGBA{0xff, 0, 0, 0xff})
	vector.StrokePath(s, &path, strokeOp, drawOp)
}

// TODO DRAW WALL
func drawEntities(s *ebiten.Image, c Component) {
	if c.visible != nil &&
		c.name != nil {
		if *c.name == "particle" {
			vector.FillCircle(s, float32(c.position.x), float32(c.position.y),
				cellSizepx, *c.color, true)
		}
	}
	if c.borderLine != nil {
		// if *c.name == "left wall" {
		drawline(s, *c.borderLine)
		//}
	}
}

// create observation zones
func (g *Game) Observe() {
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
	msg1 := fmt.Sprintf("(%d, %d)", g.ecs.Entities[0].components.position.x, g.ecs.Entities[0].components.position.y)
	msg2 := fmt.Sprintf("(%d, %d)", g.ecs.Entities[1].components.position.x, g.ecs.Entities[1].components.position.y)
	g.drawLoop(screen)
	// g.debugui.Draw(screen)

	ebitenutil.DebugPrint(screen, msg1)
	ebitenutil.DebugPrintAt(screen, msg2, 0, 10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func addlineEntity(e Ecs, v bool, n string, line BorderLine) error {
	name := n
	visible := v
	lne := line
	err := e.addEntity(Component{
		name:       &name,
		borderLine: &lne,
		visible:    &visible,
	})
	return err
}

func initParticles() Ecs {
	e := Ecs{}

	for i := range 50 {
		visibility := true
		position := Pos{rand.Intn(screenWidth-5) + 5, rand.Intn(screenHeight-5) + 5}
		velocity := Velocity{
			float32(rand.Intn(3)) + 1,
			Pos{randNum(-1, 1, int64(1+i)), randNum(-1, 1, int64(3*i))},
		}
		c1 := color.RGBA{0xff, 0, 0, 0xff}
		name := "particle"
		err := e.addEntity(Component{
			name:     &name,
			visible:  &visibility,
			position: &position,
			velocity: &velocity,

			color: &c1,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	c1 := color.RGBA{0xff, 0, 0, 0xff}
	distance := 300
	vis1 := true
	pos1 := Pos{screenWidth/2 - distance/2, screenHeight / 2}
	v1 := Velocity{
		float32(3),
		Pos{1, 0},
	}
	n1 := "particle"
	err := e.addEntity(Component{
		name:     &n1,
		visible:  &vis1,
		position: &pos1,
		velocity: &v1,
		color:    &c1,
	})
	if err != nil {
		log.Fatal(err)
	}
	vis2 := true
	pos2 := Pos{screenWidth/2 + distance/2, screenHeight / 2}
	v2 := Velocity{
		float32(3),
		Pos{-1, 0},
	}
	n2 := "particle"

	c2 := color.RGBA{0, 0, 0xff, 0xff}
	err = e.addEntity(Component{
		name:     &n2,
		visible:  &vis2,
		position: &pos2,
		velocity: &v2,
		color:    &c2,
	})
	if err != nil {
		log.Fatal(err)
	}

	name := "left wall"
	line := BorderLine{
		Pos{0, 0},
		Pos{0, screenHeight},
	}
	visibility := true
	e.addEntity(Component{
		name:       &name,
		borderLine: &line,
		visible:    &visibility,
	})

	name2 := "top wall"
	line2 := BorderLine{
		Pos{0, 0},
		Pos{screenWidth, 0},
	}
	visibility2 := true
	e.addEntity(Component{
		name:       &name2,
		borderLine: &line2,
		visible:    &visibility2,
	})
	name3 := "right wall"
	line3 := BorderLine{
		Pos{screenWidth, 0},
		Pos{screenWidth, screenHeight},
	}
	visibility3 := true
	e.addEntity(Component{
		name:       &name3,
		borderLine: &line3,
		visible:    &visibility3,
	})
	name4 := "bottom wall"
	line4 := BorderLine{
		Pos{0, screenHeight},
		Pos{screenWidth, screenHeight},
	}
	visibility4 := true
	e.addEntity(Component{
		name:       &name4,
		borderLine: &line4,
		visible:    &visibility4,
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
