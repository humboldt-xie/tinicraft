package main

import (
	"fmt"
	"time"

	_ "image/png"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/icexin/gocraft-server/proto"
)

type Game struct {
	win *glfw.Window

	player   *Player
	lx, ly   float64
	vy       float32
	prevtime float64

	blockRender  *BlockRender
	lineRender   *LineRender
	playerRender *PlayerRender

	world   *World
	itemidx int
	item    *BlockType
	fps     FPS

	exclusiveMouse bool
	closed         bool
}

func NewGame(w, h int) (*Game, error) {
	var (
		err  error
		game *Game
	)
	game = new(Game)
	game.item = &Blocks[0]

	mainthread.Call(func() {
		win := initGL(w, h)
		win.SetMouseButtonCallback(game.onMouseButtonCallback)
		win.SetCursorPosCallback(game.onCursorPosCallback)
		win.SetFramebufferSizeCallback(game.onFrameBufferSizeCallback)
		win.SetKeyCallback(game.onKeyCallback)
		game.win = win
	})
	game.world = NewWorld()
	game.player = NewPlayer(mgl32.Vec3{0, 16, 0})
	game.blockRender, err = NewBlockRender()
	if err != nil {
		return nil, err
	}
	mainthread.Call(func() {
		game.blockRender.UpdateItem(game.item)
	})
	game.lineRender, err = NewLineRender()
	if err != nil {
		return nil, err
	}
	game.playerRender, err = NewPlayerRender()
	if err != nil {
		return nil, err
	}
	go game.blockRender.UpdateLoop()
	go game.syncPlayerLoop()
	return game, nil
}

func (g *Game) setExclusiveMouse(exclusive bool) {
	if exclusive {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	g.exclusiveMouse = exclusive
}

func (g *Game) dirtyBlock(id Vec3) {
	cid := id.Chunkid()
	g.blockRender.DirtyChunk(cid)
	neighbors := []Vec3{id.Left(), id.Right(), id.Front(), id.Back()}
	for _, neighbor := range neighbors {
		chunkid := neighbor.Chunkid()
		if chunkid != cid {
			g.blockRender.DirtyChunk(chunkid)
		}
	}
}

func (g *Game) onMouseButtonCallback(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if !g.exclusiveMouse {
		g.setExclusiveMouse(true)
		return
	}
	head := NearBlock(g.player.Pos())
	foot := head.Down()
	block, prev := g.world.HitTest(g.player.Pos(), g.player.Front())
	if button == glfw.MouseButton2 && action == glfw.Press {
		if prev != nil && *prev != head && *prev != foot {
			g.world.UpdateBlock(*prev, NewBlock(g.item.Type))
			g.dirtyBlock(*prev)
			go ClientUpdateBlock(*prev, NewBlock(g.item.Type))
		}
	}
	if button == glfw.MouseButton1 && action == glfw.Press {
		if block != nil {
			tblock := g.world.Block(*block)
			if tblock != nil {
				tblock.Life -= 40
			}
			if tblock == nil || tblock.Life <= 0 {
				g.world.UpdateBlock(*block, NewBlock(typeAir))
				go ClientUpdateBlock(*block, NewBlock(typeAir))
			} else if tblock != nil {
				g.world.UpdateBlock(*block, tblock)
			}
			g.dirtyBlock(*block)
		}
	}
}

func (g *Game) onFrameBufferSizeCallback(window *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (g *Game) onCursorPosCallback(win *glfw.Window, xpos float64, ypos float64) {
	if !g.exclusiveMouse {
		return
	}
	if g.lx == 0 && g.ly == 0 {
		g.lx, g.ly = xpos, ypos
		return
	}
	dx, dy := xpos-g.lx, g.ly-ypos
	g.lx, g.ly = xpos, ypos
	g.player.ChangeAngle(float32(dx), float32(dy))
}

func (g *Game) onKeyCallback(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}
	switch key {
	case glfw.KeyTab:
		g.player.FlipFlying()
	case glfw.KeySpace:
		block := g.CurrentBlockid()
		if g.world.HasBlock(Vec3{block.X, block.Y - 2, block.Z}) {
			g.vy = 8
		}
	case glfw.KeyN:
		pos := g.player.Pos()
		PlayerID += 1
		g.playerRender.UpdateOrAdd(PlayerID, proto.PlayerState{
			X:  pos.X() + 1.0,
			Y:  pos.Y(),
			Z:  pos.Z() + 1.0,
			Rx: 5,
			Ry: 0,
		}, true)
	case glfw.KeyE:
		g.itemidx = (1 + g.itemidx) % len(Blocks)
		g.item = &Blocks[g.itemidx]
		g.blockRender.UpdateItem(g.item)
	case glfw.KeyR:
		g.itemidx--
		if g.itemidx < 0 {
			g.itemidx = len(Blocks) - 1
		}
		g.item = &Blocks[g.itemidx]
		g.blockRender.UpdateItem(g.item)
	}
}

func (g *Game) handleKeyInput(dt float64) {
	speed := float32(0.1)
	if g.player.flying {
		speed = 0.1
	}
	if g.win.GetKey(glfw.KeyEscape) == glfw.Press {
		g.setExclusiveMouse(false)
	}
	if g.win.GetKey(glfw.KeyW) == glfw.Press {
		g.player.Move(MoveForward, speed)
	}
	if g.win.GetKey(glfw.KeyS) == glfw.Press {
		g.player.Move(MoveBackward, speed)
	}
	if g.win.GetKey(glfw.KeyA) == glfw.Press {
		g.player.Move(MoveLeft, speed)
	}
	if g.win.GetKey(glfw.KeyD) == glfw.Press {
		g.player.Move(MoveRight, speed)
	}
	pos := g.player.Pos()
	stop := false
	if !g.player.Flying() {
		g.vy -= float32(dt * 20)
		if g.vy < -50 {
			g.vy = -50
		}
		pos = mgl32.Vec3{pos.X(), pos.Y() + g.vy*float32(dt), pos.Z()}
	}

	pos, stop = g.world.Collide(pos)
	if stop {
		if g.vy > -5 {
			g.vy = 0
		} else if g.vy < -5 {
			g.vy = -g.vy * 0.1
		}
	}
	g.player.SetPos(pos)

	//g.world.Generate(Vec3{int(round(pos.X())), int(round(pos.Y())), int(round(pos.Z()))})
}

func (g *Game) CurrentBlockid() Vec3 {
	pos := g.player.Pos()
	return NearBlock(pos)
}

func (g *Game) ShouldClose() bool {
	return g.closed
}

func (g *Game) renderStat() {
	g.fps.Update()
	p := g.player.Pos()
	cid := NearBlock(p).Chunkid()
	blockPos, _ := g.world.HitTest(g.player.Pos(), g.player.Front())

	life := 0
	if blockPos != nil {
		block := g.world.Block(*blockPos)
		if block != nil {
			life = block.Life
		}
	}
	stat := g.blockRender.Stat()
	title := fmt.Sprintf("[%.2f %.2f %.2f] %v [%d/%d %d] %d %d/100", p.X(), p.Y(), p.Z(),
		cid, stat.RendingChunks, stat.CacheChunks, stat.Faces, g.fps.Fps(), life)
	g.win.SetTitle(title)
}

func (g *Game) syncPlayerLoop() {
	tick := time.NewTicker(time.Second / 10)
	for range tick.C {
		ClientUpdatePlayerState(g.player.State())
	}
}

func (g *Game) Update() {
	/*pos := g.player.Pos()
	g.playerRender.UpdateOrAdd(1, proto.PlayerState{
		X:  pos.X() + 1.0,
		Y:  pos.Y(),
		Z:  pos.Z() + 1.0,
		Rx: 5,
		Ry: 0,
	})*/
	mainthread.Call(func() {
		var dt float64
		now := glfw.GetTime()
		dt = now - g.prevtime
		g.prevtime = now
		if dt > 0.02 {
			dt = 0.02
		}

		g.handleKeyInput(dt)

		gl.ClearColor(0.57, 0.71, 0.77, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		g.blockRender.Draw()
		g.lineRender.Draw()
		g.playerRender.Draw()
		g.renderStat()

		g.win.SwapBuffers()
		glfw.PollEvents()
		g.closed = g.win.ShouldClose()
	})
}
