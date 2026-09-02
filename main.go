package main

import (
    "github.com/oakmound/oak/v4"
    "github.com/oakmound/oak/v4/scene"
)

func main() {
    oak.AddScene("firstScene", scene.Scene{
        Start: func(*scene.Context) {
            // ... draw entities, bind callbacks ... 
        }, 
    })
    oak.Init("firstScene")
}
