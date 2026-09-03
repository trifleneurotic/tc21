package main

import (
	"image/color"
	"path/filepath"

	"github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/collision"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/key"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/render/mod"
	"github.com/oakmound/oak/v4/scene"
)

const GridSize = 4.0

const (
	SchoonerLabel   collision.Label = 1
	MorgLabel       collision.Label = 2
	TumbleweedLabel collision.Label = 3
	BulletLabel     collision.Label = 4
)

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
			var bulletAlive bool
			var bullet *entities.Entity
			var bulletSprite *render.Sprite
			var lockedBulletDirection float32

			schoonerSprite, err := render.LoadSprite(filepath.Join("assets/images/schooner1.png"))
			if err != nil {
				panic(err)
			}
			schooner := entities.New(ctx,
				entities.WithRenderable(schoonerSprite),
				entities.WithPosition(floatgeom.Point2{64, 64}),
				entities.WithLabel(SchoonerLabel),
			)
			render.Draw(schoonerSprite)

			bulletSprite = render.NewColorBox(8, 8, color.RGBA{R: 255, A: 255})

			event.Bind(ctx, event.Enter, schooner, func(c *entities.Entity, ev event.EnterPayload) event.Response {

				// Move left and right with A and D

				if oak.IsDown(key.A) {
					if !oak.IsDown(key.W) && !oak.IsDown(key.S) && !oak.IsDown(key.D) {
						heldLeft, _ := oak.IsHeld(key.A)
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

				}

				if oak.IsDown(key.D) {
					if !oak.IsDown(key.S) && !oak.IsDown(key.W) && !oak.IsDown(key.A) {
						heldRight, _ := oak.IsHeld(key.D)
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

				}

				if oak.IsDown(key.S) {
					if !oak.IsDown(key.W) && !oak.IsDown(key.A) && !oak.IsDown(key.D) {
						heldDown, _ := oak.IsHeld(key.S)
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

				}

				if oak.IsDown(key.W) {
					if !oak.IsDown(key.S) && !oak.IsDown(key.A) && !oak.IsDown(key.D) {
						heldUp, _ := oak.IsHeld(key.W)
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

				}

				if oak.IsDown(key.Spacebar) {
					if !bulletAlive {
						bullet = entities.New(ctx,
							entities.WithRenderable(bulletSprite),
							entities.WithLabel(BulletLabel),
							entities.WithPosition(floatgeom.Point2{schooner.X() + 16, schooner.Y() + 16}),
						)
						lockedBulletDirection = currentRotation
						render.Draw(bulletSprite)
						bulletAlive = true
					}

				}

				if bullet != nil {
					newBulletX := bullet.X()
					newBulletY := bullet.Y()

					if lockedBulletDirection == 0.0 {
						newBulletY -= 10
					} else if lockedBulletDirection == 90.0 {
						newBulletX += 10
					} else if lockedBulletDirection == 180.0 {
						newBulletY += 10
					} else if lockedBulletDirection == 270.0 {
						newBulletX -= 10
					}
					bullet.SetPos(floatgeom.Point2{newBulletX, newBulletY})
					render.Draw(bulletSprite)

					if newBulletX < 0 || newBulletX > 800 || newBulletY < 0 || newBulletY > 600 {
						bulletSprite.Undraw()
						if bullet != nil {
							bullet = nil
						}
						bulletAlive = false
					}
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
