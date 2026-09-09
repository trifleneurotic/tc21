package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"path/filepath"
	"time"

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
	SaguaroLabel    collision.Label = 20 + iota
)

type Direction int

const (
	Up Direction = iota + 5
	Down
	Left
	Right
	UpLeft
	UpRight
	DownLeft
	DownRight
)

type AdjacentResult struct {
	Space *collision.Space
	CID   event.CallerID
}

func ProcessNeighbors(mySpace *collision.Space, myTree *collision.Tree) {
	neighbors := CheckAdjacent(mySpace, myTree, Up)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, Down)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, Left)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, Right)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, UpLeft)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, UpRight)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, DownLeft)...)
	neighbors = append(neighbors, CheckAdjacent(mySpace, myTree, DownRight)...)

	for _, neighbor := range neighbors {
		// 1. Look up the core entity using Oak's event handler registry
		ent := event.DefaultCallerMap.GetEntity(neighbor.CID)
		if ent == nil {
			continue
		}

		// 2. Type-assert to your custom game struct
		if ent, ok := ent.(*entities.Entity); ok {
			// You now have access to your custom struct fields!
			print(ent.CallerID)
		}
	}
}

// CheckAdjacent looks for a non-overlapping space directly touching the current space
// in the specified cardinal direction. Returns the adjacent space found, or nil.
func CheckAdjacent(mySpace *collision.Space, myTree *collision.Tree, dir Direction) []AdjacentResult {
	if mySpace == nil {
		return nil
	}

	// 1. Get current boundaries
	x, y := mySpace.X(), mySpace.Y()
	w, h := mySpace.W(), mySpace.H()

	var scanSpace *collision.Space

	// 2. Project a 1-pixel wide/tall checking box right outside the border
	switch dir {
	case Up:
		// Directly above the top border
		scanSpace = collision.NewUnassignedSpace(x, y-1, w, 1)
	case Down:
		// Directly below the bottom border
		scanSpace = collision.NewUnassignedSpace(x, y+h, w, 1)
	case Left:
		// Directly left of the left border
		scanSpace = collision.NewUnassignedSpace(x-1, y, 1, h)
	case Right:
		// Directly right of the right border
		scanSpace = collision.NewUnassignedSpace(x+w, y, 1, h)
	// --- Diagonal Directions (1x1 pixel sensor placed corner-adjacent) ---
	case UpLeft:
		// 1 pixel up, 1 pixel left from the top-left corner
		scanSpace = collision.NewUnassignedSpace(x-1, y-1, 1, 1)
	case UpRight:
		// 1 pixel up, 1 pixel right from the top-right corner
		scanSpace = collision.NewUnassignedSpace(x+w, y-1, 1, 1)
	case DownLeft:
		// 1 pixel down, 1 pixel left from the bottom-left corner
		scanSpace = collision.NewUnassignedSpace(x-1, y+h, 1, 1)
	case DownRight:
		// 1 pixel down, 1 pixel right from the bottom-right corner
		scanSpace = collision.NewUnassignedSpace(x+w, y+h, 1, 1)
	}

	// 3. Query Oak's default CollisionTree
	// SearchIntersect returns all spaces overlapping our 1-pixel scan zone
	// hits := collision.DefaultTree.SearchIntersect(scanSpace.Bounds())
	hits := myTree.SearchIntersect(scanSpace.Bounds())

	var results []AdjacentResult

	for _, hit := range hits {
		// Ignore self
		if hit == mySpace {
			continue

		}

		fmt.Println(("Hit found!!!!"))
		// Found a valid, strictly adjacent neighbor!

		results = append(results, AdjacentResult{
			Space: hit,
			CID:   hit.CID,
		})
		// ProcessNeighbors(scanSpace, myTree)
		ent := event.DefaultCallerMap.GetEntity(hit.CID)
		fmt.Printf("%v\n", ent.CID())
		return results
	}

	return nil
}

// CheckAdjacentUsingSpace expands sprite A's collider by 1 pixel on all sides
// to check if it makes contact with sprite B's space.
func CheckAdjacentUsingSpace(a, b *entities.Entity) bool {
	spA := a.Space
	spB := b.Space

	// Create an expanded probe space around A (+1 pixel border)
	probe := collision.NewRect(
		spA.X()-1,
		spA.Y()-1,
		spA.W()+2,
		spA.H()+2,
	)

	return probe.Intersects(spB.Bounds())
}

func main() {
	myColor := color.RGBA{R: 194, G: 178, B: 128, A: 255}
	oak.SetColorBackground(image.NewUniform(myColor))

	oak.AddScene("firstScene", scene.Scene{
		Start: func(ctx *scene.Context) {
			sceneStartTime := time.Now()
			var oldElapsedSeconds int64
			var elapsedSeconds int64

			event.GlobalBind(event.DefaultBus, event.Enter, func(ev event.EnterPayload) event.Response {
				// Calculate total duration elapsed since the scene started
				// Convert to seconds (as a float64 for partial seconds)

				oldElapsedSeconds = elapsedSeconds
				elapsedSeconds = int64(time.Since(sceneStartTime) / time.Second)

				if oldElapsedSeconds != elapsedSeconds {
					fmt.Printf("Elapsed Scene Time: %v seconds\n", elapsedSeconds)
					fmt.Printf("Old Elapsed Time: %v seconds\n", oldElapsedSeconds)
				}

				return 0
			})

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
			var saguaroSprite *render.Sprite
			var saguaroGridXMax int
			var saguaroGridYMax int
			var saguaroCounter int = 0
			var oldSchoonerPosX float64
			var oldSchoonerPosY float64
			var bulletHitSaguaro bool
			var oldBulletX float64
			var oldBulletY float64

			saguaroGridXMax = 20
			saguaroGridYMax = 14

			saguaroGrid := make([][]int, saguaroGridXMax)
			for x := range saguaroGrid {
				saguaroGrid[x] = make([]int, saguaroGridYMax)
			}
			saguaros := []*entities.Entity{}

			saguaroSprite, err := render.LoadSprite(filepath.Join("assets/images/saguaro1.png"))

			for i := 0; i < 10; i++ {
				saguaroX := rand.IntN(saguaroGridXMax)
				saguaroY := rand.IntN(saguaroGridYMax)
				saguaroFound := false
				for !saguaroFound {
					if saguaroGrid[saguaroX][saguaroY] == 0 {
						saguaroGrid[saguaroX][saguaroY] = 1
						saguaroFound = true
					} else {
						saguaroX = rand.IntN(saguaroGridXMax)
						saguaroY = rand.IntN(saguaroGridYMax)
					}
				}
			}

			for saguaroX := 0; saguaroX < saguaroGridXMax; saguaroX++ {
				for saguaroY := 0; saguaroY < saguaroGridYMax; saguaroY++ {
					if saguaroGrid[saguaroX][saguaroY] == 1 {
						saguaroCounter++
						tmp := SaguaroLabel + collision.Label(saguaroCounter)
						saguaro := entities.New(ctx,
							entities.WithRenderable(saguaroSprite.Copy()),
							entities.WithPosition(floatgeom.Point2{float64((saguaroX + 1) * 32.0), float64((saguaroY + 1) * 32.0)}),
							entities.WithLabel(tmp),
						)

						saguaros = append(saguaros, saguaro)
						collision.NewLabeledSpace(float64((saguaroX+1)*32.0), float64((saguaroY+1)*32.0), 32, 32, tmp)
					}
				}
			}

			for _, saguaro := range saguaros {
				render.Draw(saguaro.Renderable)
			}

			schoonerSprite, err := render.LoadSprite(filepath.Join("assets/images/schooner1.png"))
			if err != nil {
				panic(err)
			}
			schooner := entities.New(ctx,
				entities.WithRenderable(schoonerSprite),
				entities.WithPosition(floatgeom.Point2{0, 0}),
				entities.WithLabel(SchoonerLabel),
			)

			render.Draw(schoonerSprite)

			bulletSprite = render.NewColorBox(8, 8, color.RGBA{R: 255, A: 255})

			for _, saguaro := range saguaros {
				collision.UpdateSpace(saguaro.X(), saguaro.Y(), 32.0, 32.0, saguaro.Space)
				if CheckAdjacent(saguaro.Space, ctx.CollisionTree, Up) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, Down) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, Left) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, Right) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, UpLeft) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, UpRight) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, DownLeft) != nil || CheckAdjacent(saguaro.Space, ctx.CollisionTree, DownRight) != nil {
					fmt.Println("I have a neighbor!!!!!!")
				}
			}

			event.Bind(ctx, event.Enter, schooner, func(c *entities.Entity, ev event.EnterPayload) event.Response {
				bulletHitSaguaro = false
				var hit *collision.Space
				var bulletHit *collision.Space

				for _, saguaro := range saguaros {

					collision.UpdateSpace(saguaro.X(), saguaro.Y(), 32.0, 32.0, saguaro.Space)

					hit = collision.HitLabel(schooner.Space, saguaro.Space.Label)

					if bulletAlive && !bulletHitSaguaro {
						collision.UpdateSpace(bullet.X(), bullet.Y(), 8.0, 8.0, bullet.Space)
						bulletHit = collision.HitLabel(saguaro.Space, bullet.Space.Label)

						if bulletHit != nil {
							collision.UpdateSpace(oldBulletX, oldBulletY, 8.0, 8.0, bullet.Space)
							fmt.Println("UNDRAWING")
							bulletSprite.Undraw()
							if bullet != nil {
								bullet = nil
							}
							bulletAlive = false
							bulletHitSaguaro = false

						}
					}

					if hit != nil {
						schooner.SetX(oldSchoonerPosX)
						schooner.SetY(oldSchoonerPosY)

					}
				}

				if oak.IsDown(key.A) {
					if !oak.IsDown(key.W) && !oak.IsDown(key.S) && !oak.IsDown(key.D) {
						heldLeft, _ := oak.IsHeld(key.A)
						if !leftPressed || heldLeft {
							if schooner.X()-GridSize >= 0 { // Prevent moving out of bounds
								oldSchoonerPosX = schooner.X()
								oldSchoonerPosY = schooner.Y()
								schooner.ShiftX(-GridSize)

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
								oldSchoonerPosX = schooner.X()
								oldSchoonerPosY = schooner.Y()

								schooner.ShiftX(GridSize)

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
								oldSchoonerPosX = schooner.X()
								oldSchoonerPosY = schooner.Y()

								schooner.ShiftY(GridSize)

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
								oldSchoonerPosX = schooner.X()
								oldSchoonerPosY = schooner.Y()

								schooner.ShiftY(-GridSize)

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

				collision.UpdateSpace(schooner.X(), schooner.Y(), 32.0, 32.0, schooner.Space)

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
					oldBulletX = bullet.X()
					oldBulletY = bullet.Y()

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
