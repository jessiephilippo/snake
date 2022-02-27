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
	8. Optimize screen rendering
	9. Refactoring/ cleanup code
*/

var (
	screen         tcell.Screen
	isGamePaused   bool
	isGameOver     bool
	debugLogString string
	snake          *Snake
	apple          *Apple
	score          int
)

const (
	snakeSymbol = 0x2588
	appleSymbol = 0x25CF

	gameFrameWidth  = 30
	gameFrameHeight = 15
	gameFrameSymbol = '‖'
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

	initScreen()
	initGameState()
	inputChan := initUserInput()

	for !isGameOver {
		handleUserInput(readInput(inputChan))
		updateState()
		drawState()

		time.Sleep(75 * time.Millisecond)
	}

	screenWidht, screenHeight := screen.Size()
	printStringCentered(screenWidht/2, screenHeight/2, "Game over!")
	printStringCentered(screenWidht/2-1, screenHeight/2, fmt.Sprintf("Your score is %d....", score))
	screen.Show()
	time.Sleep(3 * time.Second)
	screen.Fini()
}

func initScreen() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)
}

func initGameState() {
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

	apple = &Apple{
		point:  &Point{row: 10, col: 10},
		symbol: appleSymbol,
	}
}

func initUserInput() chan string {
	inputChan := make(chan string)
	go func() {
		for {
			switch ev := screen.PollEvent().(type) {
			case *tcell.EventKey:
				inputChan <- ev.Name()
			}
		}
	}()

	return inputChan
}

func readInput(inputChan chan string) string {
	var key string
	select {
	case key = <-inputChan:
	default:
		key = ""
	}

	return key
}

func handleUserInput(key string) {
	if key == "Rune[q]" {
		screen.Fini()
		os.Exit(0)
	} else if key == "Rune[w]" && snake.velRow != 1 {
		snake.velRow = -1
		snake.velCol = 0
	} else if key == "Rune[a]" && snake.velCol != 1 {
		snake.velRow = 0
		snake.velCol = -1
	} else if key == "Rune[s]" && snake.velRow != -1 {
		snake.velRow = 1
		snake.velCol = 0
	} else if key == "Rune[d]" && snake.velCol != -1 {
		snake.velRow = 0
		snake.velCol = 1
	} else if key == "Rune[p]" {
		isGamePaused = true
	}

}

func updateState() {
	if isGamePaused {
		return
	}

	updateSnake()
	updateApple()

}

func updateSnake() {
	head := getSnakeHead()
	snake.parts = append(snake.parts, &Point{
		row: head.row + snake.velRow,
		col: head.col + snake.velCol,
	})

	if !appleIsInsideSnake() {
		// delete first element
		snake.parts = snake.parts[1:]
	} else {
		score++
	}

	if isSnakeHittingWall() || isSnakeEatingItself() {
		isGameOver = true
	}
}

func updateApple() {
	for appleIsInsideSnake() {
		apple.point.row, apple.point.col = rand.Intn(gameFrameHeight), rand.Intn(gameFrameWidth)
	}
}

func appleIsInsideSnake() bool {
	for _, p := range snake.parts {
		if p.row == apple.point.row && p.col == apple.point.col {
			return true
		}
	}

	return false
}

func isSnakeHittingWall() bool {
	head := getSnakeHead()
	return head.row < 0 ||
		head.row >= gameFrameHeight ||
		head.col < 0 ||
		head.col >= gameFrameWidth
}

func isSnakeEatingItself() bool {
	head := getSnakeHead()
	for _, p := range snake.parts[:getSnakeHeadIndex()] {
		if p.row == head.row && p.col == head.col {
			return true
		}
	}
	return false
}

func getSnakeHeadIndex() int {
	return len(snake.parts) - 1
}

func getSnakeHead() *Point {
	return snake.parts[len(snake.parts)-1]
}
func drawState() {
	if isGamePaused {
		return
	}

	screen.Clear()

	printString(0, 0, debugLogString)
	printGameFrame()

	printSnake()
	printApple()

	screen.Show()
}

func printString(row, col int, str string) {
	for _, c := range str {
		printFilledRect(row, col, 1, 1, c)
		col += 1
	}
}

func printFilledRect(row, col, width, height int, ch rune) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			screen.SetContent(col+c, row+r, ch, nil, tcell.StyleDefault)
		}
	}
}

func printStringCentered(row, col int, str string) {
	col = col - len(str)/2
	printString(row, col, str)
}

func printGameFrame() {
	gameFrameTopLeftRow, gameFrameTopLeftCol := getFrameTopLeft()
	printUnFilledRect(gameFrameTopLeftRow-1, gameFrameTopLeftCol-1, gameFrameWidth+2, gameFrameHeight+2, gameFrameSymbol)
}

func printUnFilledRect(row, col, width, height int, ch rune) {
	// print first row
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row, ch, nil, tcell.StyleDefault)
	}

	// for each row
	// 		print first, last col
	for r := 1; r < height-1; r++ {
		screen.SetContent(col, row+r, ch, nil, tcell.StyleDefault)
		screen.SetContent(col+width-1, row+r, ch, nil, tcell.StyleDefault)

	}

	// print last row
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row+height-1, ch, nil, tcell.StyleDefault)
	}
}

func printSnake() {
	for _, p := range snake.parts {
		printFilledRectInGameFrame(p.row, p.col, 1, 1, snake.symbol)
	}
}

func printApple() {
	printFilledRectInGameFrame(apple.point.row, apple.point.col, 1, 1, apple.symbol)
}

func printFilledRectInGameFrame(row, col, widht, height int, ch rune) {
	r, c := getFrameTopLeft()
	printFilledRect(row+r, col+c, widht, height, ch)
}

func getFrameTopLeft() (int, int) {
	screenWidth, screenHeight := screen.Size()
	return screenHeight/2 - gameFrameHeight/2, screenWidth/2 - gameFrameWidth/2
}
