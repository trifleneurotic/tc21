package main

import (
	"path/filepath"

	"github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/key"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/render/mod"
	"github.com/oakmound/oak/v4/scene"
)

const GridSize = 4.0

var keyDown bool

func main() {
	oak.AddScene("firstScene", scene.Scene{
		Start: func(ctx *scene.Context) {
			var leftPressed bool
			var rightPressed bool
			var downPressed bool
			var upPressed bool
			var leftRotated bool
			var rightRotated bool
			var upRotated bool
			var downRotated bool
			var currentRotation float32 = 0.0

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
				if !keyDown {
					if oak.IsDown(key.A) {
						heldLeft, _ := oak.IsHeld(key.A)
						keyDown = true
						if !leftPressed || heldLeft {
							if schooner.X()-GridSize >= 0 { // Prevent moving out of bounds
								schooner.SetX(schooner.X() - GridSize)
								leftPressed = true
								if !leftRotated && currentRotation != 270.0 {
									oldRotation := currentRotation
									currentRotation = 270.0
									var rotateVal float32
									if oldRotation == 0.0 {
										rotateVal = 90.0
									} else if oldRotation == 180.0 {
										rotateVal = -90.0
									} else {
										rotateVal = 180.0
									}
									schoonerSprite = schoonerSprite.Modify(mod.Rotate(rotateVal)).(*render.Sprite)
									rightRotated = true
								}
							}
						}
					}
				} else {
					leftPressed = false
					leftRotated = false
					keyDown = false
				}

				if !keyDown {
					if oak.IsDown(key.D) {
						heldRight, _ := oak.IsHeld(key.D)
						keyDown = true
						if !rightPressed || heldRight {
							if schooner.X()+GridSize <= 768 { // Prevent moving out of bounds
								schooner.SetX(schooner.X() + GridSize)
								rightPressed = true
								if !rightRotated && currentRotation != 90.0 {
									oldRotation := currentRotation
									currentRotation = 90.0
									var rotateVal float32
									if oldRotation == 0.0 {
										rotateVal = -90.0
									} else if oldRotation == 180.0 {
										rotateVal = 90.0
									} else {
										rotateVal = 180.0
									}
									schoonerSprite = schoonerSprite.Modify(mod.Rotate(rotateVal)).(*render.Sprite)
									rightRotated = true
								}
							}
						}
					}
				} else {
					rightPressed = false
					rightRotated = false
					keyDown = false
				}

				if !keyDown {
					if oak.IsDown(key.S) {
						heldDown, _ := oak.IsHeld(key.S)
						keyDown = true
						if !downPressed || heldDown {
							if schooner.Y()+GridSize <= 568 { // Prevent moving out of bounds
								schooner.SetY(schooner.Y() + GridSize)
								downPressed = true
								if !downRotated && currentRotation != 180.0 {
									oldRotation := currentRotation
									currentRotation = 180.0
									var rotateVal float32
									if oldRotation == 90.0 {
										rotateVal = -90.0
									} else if oldRotation == 270.0 {
										rotateVal = 90.0
									} else {
										rotateVal = 180.0
									}
									schoonerSprite = schoonerSprite.Modify(mod.Rotate(rotateVal)).(*render.Sprite)
									downRotated = true
								}
							}
						}
					}
				} else {
					downPressed = false
					downRotated = false
					keyDown = false
				}

				if !keyDown {
					if oak.IsDown(key.W) {
						heldUp, _ := oak.IsHeld(key.W)
						keyDown = true
						if !upPressed || heldUp {
							if schooner.Y()-GridSize >= 0 { // Prevent moving out of bounds
								schooner.SetY(schooner.Y() - GridSize)
								upPressed = true
								if !upRotated && currentRotation != 0.0 {
									oldRotation := currentRotation
									currentRotation = 0.0
									var rotateVal float32
									if oldRotation == 90.0 {
										rotateVal = 90.0
									} else if oldRotation == 270.0 {
										rotateVal = -90.0
									} else {
										rotateVal = 180.0
									}
									schoonerSprite = schoonerSprite.Modify(mod.Rotate(rotateVal)).(*render.Sprite)
									upRotated = true
								}
							}
						}
					}
				} else {
					upPressed = false
					upRotated = false
					keyDown = false
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
