package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

/*
			GAME PLAN
	DONE - 1. Print game frame
	DONE - 2. Setup game objects - the data types
	DONE - 3. Initalize game objects & draw them
	DONE - 4. Handle user input for moving the snake
	DONE - 5. Eating apples and regenerate the apples
	DONE - 6. Hitting the walls & game over
	DONE - 7. Snake eating itsels
	DONE - 8. Optimize screen rendering
	DONE - 9. Refactoring/ cleanup code
*/

var (
	screen        tcell.Screen
	isGamePaused  bool
	isGameOver    bool
	snake         *Snake
	apple         *Apple
	score         int
	pointsToClear []*Point
)

const (
	snakeSymbol = 0x2B24
	appleSymbol = 0x25CF

	gameFrameWidth  = 60
	gameFrameHeight = 25
	gameFrameSymbol = '█'
)

type Point struct {
	row, col int
}

type Apple struct {
	point  *Point
	symbol rune
}

type Snake struct {
	parts          []*Point
	velRow, velCol int
	symbol         rune
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initalize screen
	initScreen()
	// Initialize the game objects
	initGameObjects()

	screen.HideCursor()
	for !isGameOver {
		printScore()
		// Read the user input
		readUserInput()

		// Update the screen and actions given
		updateState()
		// Draw the state
		drawState()

		// Refresh rate for the screen
		time.Sleep(75 * time.Millisecond)
	}

	// For when the game is over
	endCredits(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
}

func initScreen() {
	var err error
	// Create the new screen
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Initialize the screen
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// A default style for the created screen
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	// Set the style
	screen.SetStyle(defStyle)
}

func initGameObjects() {
	// Fill in the snake struct
	snake = &Snake{
		parts: []*Point{
			{row: 9, col: 3}, // tail
			{row: 8, col: 3},
			{row: 7, col: 3},
			{row: 6, col: 3},
			{row: 5, col: 3}, // head
		},
		velRow: -1,
		velCol: 0,
		symbol: snakeSymbol,
	}

	// Fill in the apple struct
	apple = &Apple{
		point:  &Point{row: 10, col: 10},
		symbol: appleSymbol,
	}
}

func readUserInput() {
	// Go routine to read the screen events
	go func() {
		switch ev := screen.PollEvent().(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'q' {
				screen.Fini()
				os.Exit(1)
			} else if ev.Rune() == 'w' && snake.velRow != 1 {
				snake.velRow = -1
				snake.velCol = 0
			} else if ev.Rune() == 'a' && snake.velCol != 1 {
				snake.velRow = 0
				snake.velCol = -1
			} else if ev.Rune() == 's' && snake.velRow != -1 {
				snake.velRow = 1
				snake.velCol = 0
			} else if ev.Rune() == 'd' && snake.velCol != -1 {
				snake.velRow = 0
				snake.velCol = 1
			} else if ev.Rune() == 'p' {
				isGamePaused = !isGamePaused
			}
		}
	}()
}

func updateState() {
	// If it is paused then don't do anything :)
	if isGamePaused {
		return
	}

	// Update the snake model
	updateSnake()
	// Update the apple and randomize the spawn
	updateApple()
}

func drawState() {
	// If the game is paused then do nothing :)
	if isGamePaused {
		return
	}

	clearScreen()
	// retrieve the top left game frame
	// Print the game frame
	drawGameFrame()

	// Print the snake
	drawSnake()
	// Print the apple
	drawApple()

	// Show all the stuff on screen :D
	screen.Show()
}

func clearScreen() {
	for _, p := range pointsToClear {
		printFilledRectInGameFrame(p.row, p.col, 1, 1, ' ', tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack))
	}

	// Clear the buffer
	pointsToClear = []*Point{}

}

func endCredits(style tcell.Style) {
	screenWidht, screenHeight := screen.Size()
	printStringCentered(screenWidht/2-1, screenHeight/2-1, "Game over!", style)
	printStringCentered(screenWidht/2, screenHeight/2, fmt.Sprintf("Your score = %d", score), style)

	// Show the prints
	screen.Show()

	// Shows the prints for a x amount of seconds
	time.Sleep(3 * time.Second)

	// Finish the screne <- aka terminate
	screen.Fini()
}

func updateSnake() {
	// Retrieve the head of the snake
	head := getSnakeHead()
	snake.parts = append(snake.parts, &Point{
		row: head.row + snake.velRow,
		col: head.col + snake.velCol,
	})

	if !appleIsInsideSnake() {
		// delete first element
		snake.parts = snake.parts[1:]
	} else {
		// Update the score
		score++
	}

	// If the snake is hitting one of the walls or if it touched itself.
	// Then it should be game over
	if isSnakeHittingWall() || isSnakeEatingItself() {
		isGameOver = true
	}
}

func updateApple() {
	for appleIsInsideSnake() {
		// Randomize the spawn of the apple
		apple.point.row, apple.point.col = rand.Intn(gameFrameHeight), rand.Intn(gameFrameWidth)
	}
}

func appleIsInsideSnake() bool {
	// Make sure that if the apple spawns inside the snake to give it a random spawn
	// This is more for late game
	for _, p := range snake.parts {
		if p.row == apple.point.row && p.col == apple.point.col {
			return true
		}
	}

	return false
}

func isSnakeHittingWall() bool {
	// Get the position of the snakes head
	head := getSnakeHead()
	// Do some logic to determine if the snake has touched a wall
	return head.row < 0 ||
		head.row >= gameFrameHeight ||
		head.col < 0 ||
		head.col >= gameFrameWidth
}

func isSnakeEatingItself() bool {
	// Get the position of the snakes head
	head := getSnakeHead()
	// Do some logic to determine if the snake has touched himself
	for _, p := range snake.parts[:len(snake.parts)-1] {
		if p.row == head.row && p.col == head.col {
			return true
		}
	}
	return false
}

func getSnakeHead() *Point {
	// retrieve the position of the snakes head
	return snake.parts[len(snake.parts)-1]
}

func drawGameFrame() {
	gameFrameRow, gameFrameCol := getFrame()
	row, col, width, height := gameFrameRow-1, gameFrameCol-1, gameFrameWidth+2, gameFrameHeight+2
	ch := gameFrameSymbol

	// Set a default style for the frame
	style := tcell.StyleDefault.Foreground(tcell.ColorBlueViolet)
	// print left frame
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row, ch, nil, style)
	}

	// Print top and bottom frame
	for r := 1; r < height-1; r++ {
		screen.SetContent(col, row+r, ch, nil, style)
		screen.SetContent(col+width-1, row+r, ch, nil, style)

	}

	// print right frame
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row+height-1, ch, nil, style)
	}
}

func drawSnake() {
	// iterate over the snake.parts <- to get the points
	for _, p := range snake.parts {
		// Print the snake
		printFilledRectInGameFrame(p.row, p.col, 1, 1, snake.symbol, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen))
		pointsToClear = append(pointsToClear, p)
	}
}

func drawApple() {
	// Print the apple
	printFilledRectInGameFrame(apple.point.row, apple.point.col, 1, 1, apple.symbol, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed))
	pointsToClear = append(pointsToClear, apple.point)
}

func printFilledRect(row, col, width, height int, ch rune, style tcell.Style) {
	// This prints the filled rectangle
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			// Set the content to show on screen
			screen.SetContent(col+c, row+r, ch, nil, style)
		}
	}
}

func printStringCentered(col, row int, str string, style tcell.Style) {
	for _, c := range str {
		screen.SetContent(col-len(str)/2, row, c, nil, style)
		col += 1
	}
}

func getFrame() (int, int) {
	// Retrieve the screen width and height
	screenWidth, screenHeight := screen.Size()
	return screenHeight/2 - gameFrameHeight/2, screenWidth/2 - gameFrameWidth/2
}

func printFilledRectInGameFrame(row, col, widht, height int, ch rune, style tcell.Style) {
	r, c := getFrame()
	printFilledRect(row+r, col+c, widht, height, ch, style)
}

func printScore() {
	screenWidth, screenHeight := getFrame()
	for _, c := range fmt.Sprintf("Current score: %d", score) {
		screen.SetContent(screenHeight, screenWidth-2, c, nil, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
		screenHeight += 1
	}
	screen.Show()
}
