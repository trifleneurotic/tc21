package main

import (
	"fmt"
	"path/filepath"

	"github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/key"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/scene"
)

const GridSize = 4.0

func main() {
	oak.AddScene("firstScene", scene.Scene{
		Start: func(ctx *scene.Context) {
			var leftPressed bool
			var rightPressed bool
			var downPressed bool
			var upPressed bool
			schoonerSprite, err := render.LoadSprite(filepath.Join("assets/images/schooner1.png"))
			if err != nil {
				panic(err)
			}
			schooner := entities.New(ctx,
				entities.WithRenderable(schoonerSprite),
				entities.WithPosition(floatgeom.Point2{64, 64}),
			)
			render.Draw(schoonerSprite)

			event.Bind(ctx, event.Enter, schooner, func(c *entities.Entity, ev event.EnterPayload) event.Response {

				// Move left and right with A and D
				heldLeft, _ := oak.IsHeld(key.A)
				if oak.IsDown(key.A) {
					if !leftPressed || heldLeft {
						if schooner.X()-GridSize >= 0 { // Prevent moving out of bounds
							schooner.SetX(schooner.X() - GridSize)
							fmt.Println(schooner.X())
							leftPressed = true
						}
					}
				} else {
					leftPressed = false
				}

				heldRight, _ := oak.IsHeld(key.D)
				if oak.IsDown(key.D) {
					if !rightPressed || heldRight {
						if schooner.X()+GridSize <= 768 { // Prevent moving out of bounds
							schooner.SetX(schooner.X() + GridSize)
							fmt.Println(schooner.X())
							rightPressed = true
						}
					}
				} else {
					rightPressed = false
				}

				heldDown, _ := oak.IsHeld(key.S)
				if oak.IsDown(key.S) {
					if !downPressed || heldDown {
						if schooner.Y()+GridSize <= 568 { // Prevent moving out of bounds
							schooner.SetY(schooner.Y() + GridSize)
							fmt.Println(schooner.Y())
							downPressed = true
						}
					}
				} else {
					downPressed = false
				}

				heldUp, _ := oak.IsHeld(key.W)
				if oak.IsDown(key.W) {
					if !upPressed || heldUp {
						if schooner.Y()-GridSize >= 0 { // Prevent moving out of bounds
							schooner.SetY(schooner.Y() - GridSize)
							fmt.Println(schooner.Y())
							upPressed = true
						}
					}
				} else {
					upPressed = false
				}

				return 0
			})

		},
	})

	oak.Init("firstScene", func(c oak.Config) (oak.Config, error) {
		c.Screen.Width = 800
		c.Screen.Height = 600
		c.Screen.Scale = 1
		c.Title = "Tombstone City: 21st Century"
		return c, nil
	})
}
