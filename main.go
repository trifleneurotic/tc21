package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"path/filepath"
	"slices"
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
	MorgLabel       collision.Label = 50 + iota
	TumbleweedLabel collision.Label = 3
	BulletLabel     collision.Label = 4
	SaguaroLabel    collision.Label = 20 + iota
	PairLabel       collision.Label = 80 + iota
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

var (
	enemySpawnEvent  = event.RegisterEvent[*scene.Context]()
	enemyActionReady = event.RegisterEvent[*entities.Entity]()
	morgs            []*entities.Entity
)

var enemyIDCounter = 30
var enemyCounter = 0
var saguaroPairs = map[int][]*entities.Entity{}

// Direction offset pairs for all 8 neighbors (Row, Column)
var directions = [][2]int{
	{-1, -1}, {-1, 0}, {-1, 1}, // Top-left, Top, Top-right
	{0, -1}, {0, 1}, // Left,      Right
	{1, -1}, {1, 0}, {1, 1}, // Bottom-left, Bottom, Bottom-right
}

func spawnEnemyWithAdHocTimer(ctx *scene.Context) {
	if len(saguaroPairs) > 0 {
		fmt.Printf("Spawning Enemy and initiating an ad-hoc 3-second timer...\n")

		keys := make([]int, 0, len(saguaroPairs))
		for k := range saguaroPairs {
			keys = append(keys, k)
		}

		randomKey := keys[rand.IntN(len(keys))]
		randomPair := saguaroPairs[randomKey]
		randomIndex := rand.IntN(2)

		spawnImpending := render.NewColorBox(32, 32, color.RGBA{255, 0, 0, 255})
		spawnDone := render.NewColorBox(32, 32, color.RGBA{0, 0, 0, 0})

		sw := render.NewSwitch("impend", map[string]render.Modifiable{
			"spawnDone":      spawnDone,
			"spawnImpending": spawnImpending,
		})

		sw.Set("spawnImpending")
		sw.SetPos(randomPair[randomIndex].X(), randomPair[randomIndex].Y())
		render.Draw(sw)

		// 2. AD-HOC TIMER: Start a non-blocking 3-second delay right now for THIS enemy
		time.AfterFunc(3*time.Second, func() {
			sw.Set("spawnDone")
			sw.SetPos(randomPair[randomIndex].X(), randomPair[randomIndex].Y())
			render.Draw(sw)

			enemyCounter++
			tmp := MorgLabel + collision.Label(enemyCounter)
			morg := entities.New(ctx,
				entities.WithLabel(tmp),
				entities.WithRenderable(render.NewColorBox(32, 32, color.RGBA{0, 255, 255, 255})),
			)
			event.DefaultBus.Trigger(enemyActionReady.UnsafeEventID, morg)

			// spawn Morg
			if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, Up) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X(), randomPair[randomIndex].Y() - 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X(), randomPair[randomIndex].Y()-32.0, 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, Down) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X(), randomPair[randomIndex].Y() + 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X(), randomPair[randomIndex].Y()+32.0, 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, Left) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() - 32.0, randomPair[randomIndex].Y()})
				collision.UpdateSpace(randomPair[randomIndex].X()-32.0, randomPair[randomIndex].Y(), 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, Right) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() + 32.0, randomPair[randomIndex].Y()})
				collision.UpdateSpace(randomPair[randomIndex].X()+32.0, randomPair[randomIndex].Y(), 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, DownLeft) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() - 32.0, randomPair[randomIndex].Y() + 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X()-32.0, randomPair[randomIndex].Y()+32.0, 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, DownRight) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() + 32.0, randomPair[randomIndex].Y() + 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X()+32.0, randomPair[randomIndex].Y()+32.0, 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, UpRight) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() + 32.0, randomPair[randomIndex].Y() - 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X()+32.0, randomPair[randomIndex].Y()-32.0, 32, 32, morg.Space)
			} else if CheckAdjacent(randomPair[randomIndex].Space, ctx.CollisionTree, UpLeft) == nil {
				morg.SetPos(floatgeom.Point2{randomPair[randomIndex].X() - 32.0, randomPair[randomIndex].Y() - 32.0})
				collision.UpdateSpace(randomPair[randomIndex].X()-32.0, randomPair[randomIndex].Y()-32.0, 32, 32, morg.Space)
			}
			morgs = append(morgs, morg)
		})
	} else {
		fmt.Println("No more pairs!!!!")
	}
}

// hasNeighbors checks if an element at (r, c) has any neighbors.
// In a fully populated grid, elements always have neighbors unless the grid is 1x1.
// This function prints the neighbors found.
func hasNeighbors(grid [][]int, r, c int) bool {
	rows := len(grid)
	if rows == 0 {
		return false
	}
	cols := len(grid[0])

	// Validate if the starting point itself is inside the grid
	if r < 0 || r >= rows || c < 0 || c >= cols {
		return false
	}

	foundNeighbor := false

	// Loop through all 8 possible directions
	for _, d := range directions {
		newRow := r + d[0]
		newCol := c + d[1]

		// Bound checking: Ensure the neighbor is inside the grid
		if newRow >= 0 && newRow < rows && newCol >= 0 && newCol < cols && (grid[newRow][newCol] == 1 || grid[newRow][newCol] == 77) {
			foundNeighbor = true
			fmt.Printf("Neighbor found at [%d][%d] with value: %d\n", newRow, newCol, grid[newRow][newCol])
		}
	}

	return foundNeighbor
}

type AdjacentResult struct {
	Space *collision.Space
	CID   event.CallerID
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

			// 1. GLOBAL BINDER: Listens for any enemy's ad-hoc timer to finish

			event.GlobalBind(ctx, enemySpawnEvent, func(c *scene.Context) event.Response {
				spawnEnemyWithAdHocTimer(ctx)
				return 0
			})

			go func() {
				ticker := time.NewTicker(13 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						event.DefaultBus.Trigger(enemySpawnEvent.UnsafeEventID, ctx)
					case <-ctx.Done():
						return
					}
				}
			}()

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
			var schooner *entities.Entity
			var bulletSprite *render.Sprite
			var lockedBulletDirection float32
			var saguaroSprite *render.Sprite
			var saguaroGridXMax int
			var saguaroGridYMax int
			var saguaroCounter int = 0
			var pairCounter int = 0
			var oldSchoonerPosX float64
			var oldSchoonerPosY float64
			var oldBulletX float64
			var oldBulletY float64

			saguaroPairs = make(map[int][]*entities.Entity)

			event.GlobalBind(ctx, enemyActionReady, func(entity *entities.Entity) event.Response {
				return 0
			})

			saguaroGridXMax = 20
			saguaroGridYMax = 14

			saguaroGrid := make([][]int, saguaroGridXMax)
			for x := range saguaroGrid {
				saguaroGrid[x] = make([]int, saguaroGridYMax)
			}
			saguaros := []*entities.Entity{}

			safeZoneXEnd := 12
			safeZoneYEnd := 10

			for safeZoneXStart := 7; safeZoneXStart <= safeZoneXEnd; safeZoneXStart++ {
				for safeZoneYStart := 5; safeZoneYStart <= safeZoneYEnd; safeZoneYStart++ {
					saguaroGrid[safeZoneXStart][safeZoneYStart] = 77
				}
			}

			saguaroSprite, err := render.LoadSprite(filepath.Join("assets/images/saguaro1.png"))

			var pairMade = false

			// make 10 saguaros, making sure that a saguaro isn't already there
			for i := 0; i < 10; i++ {
				saguaroX := rand.IntN(saguaroGridXMax)
				saguaroY := rand.IntN(saguaroGridYMax)

				saguaroFound := false
				for !saguaroFound {
					if saguaroGrid[saguaroX][saguaroY] == 0 && !hasNeighbors(saguaroGrid, saguaroX, saguaroY) {
						saguaroGrid[saguaroX][saguaroY] = 1
						saguaroFound = true
					} else {
						saguaroX = rand.IntN(saguaroGridXMax)
						saguaroY = rand.IntN(saguaroGridYMax)

					}
				}
			}

			for i, row := range saguaroGrid {
				pairMade = false
				for j, val := range row {
					// if a saguaro is there and it doesn't have neighbors....
					if val == 1 && !hasNeighbors(saguaroGrid, i, j) {
						// ....pick a direction at random....
						dirIndex := rand.IntN(8)

						pairCounter++

						saguaroGrid[i][j] = int(PairLabel) + pairCounter

						saguaroIndexX := i + (directions[dirIndex][0])
						saguaroIndexY := j + (directions[dirIndex][1])

						fmt.Printf("ij -> %v %v\n", saguaroIndexX, saguaroIndexY)

						if saguaroIndexX == -1 {
							saguaroIndexX = 1
						}

						if saguaroIndexY == -1 {
							saguaroIndexY = 1
						}

						if saguaroIndexY == 14 {
							saguaroIndexY--
						}

						if saguaroIndexX == 14 {
							saguaroIndexX--
						}

						fmt.Printf("saguaro index -> %v %v\n", saguaroIndexX, saguaroIndexY)
						// ...and put a saguaro there to make a single pair only
						saguaroGrid[saguaroIndexX][saguaroIndexY] = int(PairLabel) + pairCounter
						pairMade = true
						fmt.Printf("pair made!!!! %v\n", saguaroGrid[saguaroIndexX][saguaroIndexY])
						if pairCounter == 2 {
							break
						}
					}
				}
				if pairMade && pairCounter == 2 {
					break
				}
			}

			for saguaroX := 0; saguaroX < saguaroGridXMax; saguaroX++ {
				for saguaroY := 0; saguaroY < saguaroGridYMax; saguaroY++ {
					if saguaroGrid[saguaroX][saguaroY] == 1 || saguaroGrid[saguaroX][saguaroY] >= 80 {
						saguaroCounter++
						tmp := SaguaroLabel + collision.Label(saguaroCounter)
						saguaro := entities.New(ctx,
							entities.WithRenderable(saguaroSprite.Copy()),
							entities.WithPosition(floatgeom.Point2{float64((saguaroX + 1) * 32.0), float64((saguaroY + 1) * 32.0)}),
							entities.WithLabel(tmp),
						)

						saguaros = append(saguaros, saguaro)
						// collision.NewLabeledSpace(float64((saguaroX+1)*32.0), float64((saguaroY+1)*32.0), 32, 32, tmp)

						if saguaroGrid[saguaroX][saguaroY] >= 80 {
							fmt.Println("adding pair..........")

							saguaro.Space.Label = collision.Label(saguaroGrid[saguaroX][saguaroY])

							pairID := saguaroGrid[saguaroX][saguaroY]
							if _, ok := saguaroPairs[pairID]; !ok {
								saguaroPairs[pairID] = []*entities.Entity{}
							}
							saguaroPairs[pairID] = append(saguaroPairs[pairID], saguaro)
						}
					}
					if saguaroGrid[saguaroX][saguaroY] == 77 {
						entities.New(ctx,
							entities.WithRenderable(render.NewColorBox(32, 32, color.RGBA{255, 255, 0, 255})),
							entities.WithPosition(floatgeom.Point2{float64((saguaroX + 1) * 32.0), float64((saguaroY + 1) * 32.0)}),
						)

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
			schooner = entities.New(ctx,
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
			event.GlobalBind(ctx, event.Enter, func(ev event.EnterPayload) event.Response {
				var hit *collision.Space
				var bulletHit *collision.Space
				var idxMorg int

				idxMorg = -1

				for _, saguaro := range saguaros {

					collision.UpdateSpace(saguaro.X(), saguaro.Y(), 32.0, 32.0, saguaro.Space)

					hit = collision.HitLabel(schooner.Space, saguaro.Space.Label)

					if bulletAlive {
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

						}
					}

					if hit != nil {
						schooner.SetX(oldSchoonerPosX)
						schooner.SetY(oldSchoonerPosY)
					}

				}

				for idx, morg := range morgs {
					pt := floatgeom.Point2{morg.X(), morg.Y()}
					pt2 := floatgeom.Point2{schooner.X(), schooner.Y()}
					delta := pt2.Sub(pt).Normalize().MulConst(80.0 * ev.SinceLastFrame.Seconds())
					morg.ShiftPos(delta.X(), delta.Y())
					collision.UpdateSpace(morg.X(), morg.Y(), 32.0, 32.0, morg.Space)

					for _, saguaro := range saguaros {
						collision.UpdateSpace(saguaro.X(), saguaro.Y(), 32.0, 32.0, saguaro.Space)
						if collision.HitLabel(saguaro.Space, morg.Space.Label) != nil {
							fmt.Println("####morg hit saguaro####")
							morg.ShiftPos(-delta.X(), -delta.Y())
							collision.UpdateSpace(morg.X(), morg.Y(), 32.0, 32.0, morg.Space)
							break
						}
					}

					if bulletAlive {
						collision.UpdateSpace(bullet.X(), bullet.Y(), 8.0, 8.0, bullet.Space)
						bulletHit = collision.HitLabel(morg.Space, bullet.Space.Label)

						if bulletHit != nil {
							var pairHit bool
							var toRemove int
							for k, v := range saguaroPairs {
								newSaguaros := []*entities.Entity{}
								pt = floatgeom.Point2{v[0].X(), v[0].Y()}
								if CheckAdjacentUsingSpace(v[0], morg) || CheckAdjacentUsingSpace(v[1], morg) {
									// remove from grid
									v[0].Renderable.Undraw()
									v[1].Renderable.Undraw()

									toRemove = k
									pairHit = true

									for _, saguaro := range saguaros {
										if k != int(saguaro.Space.Label) {
											newSaguaros = append(newSaguaros, saguaro)
										}
									}

									saguaros = newSaguaros

									for saguaroX := 0; saguaroX < saguaroGridXMax; saguaroX++ {
										for saguaroY := 0; saguaroY < saguaroGridYMax; saguaroY++ {
											if saguaroGrid[saguaroX][saguaroY] >= 80 {
												saguaroGrid[saguaroX][saguaroY] = 0
											}
										}
									}
								}
							}

							collision.UpdateSpace(oldBulletX, oldBulletY, 8.0, 8.0, bullet.Space)
							fmt.Println("UNDRAWING BULLET AND REPLACING MORG")
							bulletSprite.Undraw()

							if pairHit {
								enemyCounter++
								tmp := MorgLabel + collision.Label(enemyCounter)
								morg.Renderable.Undraw()
								newMorg := entities.New(ctx,
									entities.WithLabel(tmp),
									entities.WithRenderable(render.NewColorBox(32, 32, color.RGBA{0, 255, 255, 255})),
								)
								newMorg.SetPos(floatgeom.Point2{saguaroPairs[toRemove][0].X(), saguaroPairs[toRemove][0].Y()})
								morgs = append(morgs, newMorg)
								event.DefaultBus.Trigger(enemyActionReady.UnsafeEventID, newMorg)
								delete(saguaroPairs, toRemove)
								fmt.Printf("+++++saguaroPairs length %v", len(saguaroPairs))

							} else {
								saguaroCounter++
								tmp := SaguaroLabel + collision.Label(saguaroCounter)
								s := entities.New(ctx,
									entities.WithRenderable(saguaroSprite.Copy()),
									entities.WithPosition(floatgeom.Point2{morg.X(), morg.Y()}),
									entities.WithLabel(tmp),
								)

								collision.UpdateSpace(s.X(), s.Y(), 32.0, 32.0, s.Space)

								for _, saguaro := range saguaros {
									if CheckAdjacentUsingSpace(s, saguaro) {
										fmt.Println("NNNNNNNNNNNEW PAIR MADE!")
										pairCounter++
										saguaroPairs[int(PairLabel)+pairCounter] = []*entities.Entity{}
										s.Space.Label = PairLabel + collision.Label(pairCounter)
										saguaro.Space.Label = PairLabel + collision.Label(pairCounter)
										saguaroPairs[int(PairLabel)+pairCounter] = append(saguaroPairs[int(PairLabel)+pairCounter], s, saguaro)
									}
								}
								saguaros = append(saguaros, s)
								morg.Renderable.Undraw()
								render.Draw(s.Renderable)
							}

							if bullet != nil {
								bullet = nil
							}
							bulletAlive = false
							idxMorg = idx

						}

					}
				}

				if idxMorg > -1 {
					fmt.Println("%v %v", idxMorg, len(morgs))
					morgs = slices.Delete(morgs, idxMorg, idxMorg+1)
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
