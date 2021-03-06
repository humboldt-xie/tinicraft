package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	_ "image/png"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/humboldt-xie/tinycraft/render"
	"github.com/humboldt-xie/tinycraft/world"
)

type CircleAI struct {
	g *Game
}

func (c *CircleAI) Think(p *world.Player) {
	//_, prev := game.world.HitTest(p.Pos(), p.Front())
	p.ChangeAngle(1, 0)
	//if prev == nil {
	p.Move(world.MoveForward, 0.1)
	game.BreakBlock(p)
	//}
}

type SimplePhysics struct {
	vx, vz, vy float32
}

func (sp *SimplePhysics) GetSpeed() mgl32.Vec3 {
	return mgl32.Vec3{sp.vx, sp.vy, sp.vz}
}
func (sp *SimplePhysics) Speed(a mgl32.Vec3) {
	sp.vx = a.X()
	sp.vy = a.Y()
	sp.vz = a.Z()
}

func (sp *SimplePhysics) Update(p *world.Player, dt float64) {
	from := p.Pos()
	pos := p.Pos()
	stop := false
	if !p.Flying() {
		sp.vy -= float32(dt * 10)
		if sp.vy < -30 {
			sp.vy = -30
		}
		pos = mgl32.Vec3{
			from.X() + sp.vx*float32(dt),
			from.Y() + sp.vy*float32(dt),
			from.Z() + sp.vz*float32(dt),
		}
	}

	pos, stop = game.world.Collide(from, pos)
	if stop {
		sp.vx = 0
		sp.vz = 0
		if sp.vy > -5 {
			sp.vy = 0
		} else if sp.vy < -5 {
			sp.vy = -sp.vy * 0.1
		}
	}
	p.SetPos(pos)
}

type Game struct {
	win *glfw.Window

	players  sync.Map
	player   *world.Player
	lx, ly   float64
	prevtime float64

	blockRender  *render.BlockRender
	lineRender   *render.LineRender
	playerRender *render.PlayerRender

	world        *world.World
	itemidx      int
	item         *BlockType
	fps          FPS
	fpsObject    FPS
	prevStatTime time.Time

	exclusiveMouse bool
	closed         bool
}

func NewGame(w, h int) (*Game, error) {
	var (
		err  error
		game *Game
	)
	game = new(Game)
	game.itemidx = 24
	game.item = &Blocks[24]

	mainthread.Call(func() {
		win := initGL(w, h)
		win.SetMouseButtonCallback(game.onMouseButtonCallback)
		win.SetCursorPosCallback(game.onCursorPosCallback)
		win.SetFramebufferSizeCallback(game.onFrameBufferSizeCallback)
		win.SetKeyCallback(game.onKeyCallback)
		game.win = win
	})
	game.world = world.NewWorld(*render.RenderRadius)
	game.player = world.NewPlayer(mgl32.Vec3{0, 16, 0}, nil, &SimplePhysics{})
	err = InitConfig("mods/block/config.yaml")
	if err != nil {
		panic(err)
	}

	game.blockRender, err = render.NewBlockRender(game.win, game.world, game.player)
	if err != nil {
		return nil, err
	}
	mainthread.Call(func() {
		game.blockRender.UpdateItem(game.item)
	})
	game.lineRender, err = render.NewLineRender(game.win, game.world)
	if err != nil {
		return nil, err
	}
	game.playerRender, err = render.NewPlayerRender()
	if err != nil {
		return nil, err
	}

	//game.playerRender.Add(0, game.player)
	//if client == nil {
	game.players.Store(int32(0), game.player)
	/*} else {
		game.players.Store(int32(client.ClientID), game.player)
	}*/
	go game.watchWorld()

	go game.syncPlayerLoop()
	return game, nil
}

func (g *Game) watchWorld() {
	for {
		ch := g.world.Watcher.Watch(1024)
		for {
			ev, ok := <-ch
			if !ok {
				break
			}
			log.Printf("onEvent %v", ev)
		}
	}

}

func (g *Game) setExclusiveMouse(exclusive bool) {
	if exclusive {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	g.exclusiveMouse = exclusive
}
func (g *Game) UpdateBlock(id world.Vec3, tp *world.Block) {
	g.world.UpdateBlock(id, tp)
	g.blockRender.DirtyBlock(id)
	//go ClientUpdateBlock(id, tp)
}

func (g *Game) PutBlock(player *world.Player, item *BlockType) {
	head := player.Head()
	foot := player.Foot()
	_, prev := g.world.HitTest(player.Pos(), player.Front())
	if prev != nil && *prev != head && *prev != foot {
		g.UpdateBlock(*prev, world.NewBlock(item.Type))
	}

}
func (g *Game) SelectBlock(player *world.Player) *world.Block {
	hitPos, _ := g.world.HitTest(player.Pos(), player.Front())
	if hitPos != nil {
		tblock := g.world.Block(*hitPos)
		if tblock != nil {
			return tblock
		}
		if tblock == nil {
			tblock = world.NewBlock(world.TypeAir)
			g.UpdateBlock(*hitPos, tblock)
		}
		return g.world.Block(*hitPos)
	}
	return nil
}
func (g *Game) BreakBlock(player *world.Player) {
	tblock := g.SelectBlock(player) //g.world.HitTest(player.Pos(), player.Front())
	if tblock != nil {
		id := tblock.ID
		tblock.Life -= 40
		if tblock.Life <= 0 {
			tblock = world.NewBlock(world.TypeAir)
		}
		g.UpdateBlock(id, tblock)
	}
}

func (g *Game) onMouseButtonCallback(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if !g.exclusiveMouse {
		g.setExclusiveMouse(true)
		return
	}

	if button == glfw.MouseButton2 && action == glfw.Press {
		g.PutBlock(g.player, g.item)
	}
	if button == glfw.MouseButton1 && action == glfw.Press {
		g.BreakBlock(g.player)
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
func (g *Game) Jump(delta float32) {
	block := game.CurrentBlockid()
	if game.world.HasBlock(world.Vec3{block.X, block.Y - 2, block.Z}) {
		g.player.Physics.Speed(mgl32.Vec3{0, delta, 0})
	}
}
func (g *Game) JumpFront(delta float32) {
	block := game.CurrentBlockid()
	if game.world.HasBlock(world.Vec3{block.X, block.Y - 2, block.Z}) {
		//vec:=mgl32.Vec3{0, delta, 0}
		vec := g.player.WalkFront().Mul(delta)
		vec[1] = delta
		g.player.Physics.Speed(vec)
	}
}

func (g *Game) onKeyCallback(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}
	switch key {
	case glfw.KeyTab:
		g.player.FlipFlying()
	case glfw.KeySpace:
		g.Jump(8)
	case glfw.KeyJ, glfw.KeyLeftShift:
		g.JumpFront(20)
	case glfw.KeyN:
		for i := 0; i < 100; i++ {
			pos := g.player.Pos()
			pos[0] = pos.X() + float32(50-i)
			//front := g.player.Front()
			world.PlayerID += 1
			v := g.player.Front().Mul(20)
			player := world.NewPlayer(pos, nil, &SimplePhysics{vx: v.X(), vy: 20, vz: v.Z()})
			game.players.Store(world.PlayerID, player)
			log.Printf("new player %v pos:%v", v, player.Pos())
			/*g.playerRender.UpdateOrAdd(PlayerID, PlayerState{
				X:  pos.X() + front.X() + float32(i)/5,
				Y:  pos.Y() + front.Y(),
				Z:  pos.Z() + front.Z() + float32(i)/5,
				Rx: g.player.State().Rx,
				Ry: g.player.State().Ry,
			}, true)*/
		}
	case glfw.KeyE:
		g.itemidx = (1 + g.itemidx) % len(Blocks)
		log.Printf("item idx %d", g.itemidx)
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
	speed := float32(3) * float32(dt)
	if g.player.Flying() {
		speed = 3 * float32(dt)
	}
	if g.win.GetKey(glfw.KeyEscape) == glfw.Press {
		g.setExclusiveMouse(false)
	}
	if g.win.GetKey(glfw.KeyW) == glfw.Press {
		g.player.Move(world.MoveForward, speed)
	}
	if g.win.GetKey(glfw.KeyS) == glfw.Press {
		g.player.Move(world.MoveBackward, speed)
	}
	if g.win.GetKey(glfw.KeyA) == glfw.Press {
		g.player.Move(world.MoveLeft, speed)
	}
	if g.win.GetKey(glfw.KeyD) == glfw.Press {
		g.player.Move(world.MoveRight, speed)
	}
}

func (g *Game) CurrentBlockid() world.Vec3 {
	pos := g.player.Pos()
	return world.NearBlock(pos)
}

func (g *Game) ShouldClose() bool {
	return g.closed
}

func (g *Game) renderStat() {
	now := time.Now()
	if now.Sub(g.prevStatTime) < time.Second {
		return
	}
	g.prevStatTime = now
	p := g.player.Pos()
	cid := world.NearBlock(p).Chunkid()
	c := g.world.Chunk(cid)

	life := 0
	blockType := -1
	show := render.FaceFilter{}
	block := g.SelectBlock(g.player)
	if block != nil {
		life = block.Life
		show = render.ShowFaces(g.world, block.ID)
		blockType = block.Type
	}
	stat := g.blockRender.Stat()
	stats := [][]interface{}{
		{"pos:[%.2f,%.2f,%.2f]", p.X(), p.Y(), p.Z()},
		{"fps: %3d object fps: %3d", g.fps.Fps(), g.fpsObject.Fps()},
		{"cid: %v   v:%d", cid, c.V()},
		{"rending chunks:%.5d cache: %.5d", stat.RendingChunks, stat.CacheChunks},
		{"faces: %d", stat.Faces},
		{"life: %v", life},
		{"show: %v", show},
		{"type: %d", blockType},
	}
	title := ""
	for _, v := range stats {
		title += fmt.Sprintf(v[0].(string), v[1:]...) + "\n"
	}

	g.blockRender.UpdateText(title)

	g.win.SetTitle(fmt.Sprintf("fps:%d", g.fps.Fps()))
}

func (g *Game) syncPlayerLoop() {
	/*tick := time.NewTicker(time.Second / 10)
	for range tick.C {
		ClientUpdatePlayerState(g.player.State())
	}*/
}
func (g *Game) UpdateObject() {
	fps := 200
	delta := time.Second / time.Duration(fps)
	prev := time.Now()
	dt := 1.0 / float64(fps)
	for !game.ShouldClose() {
		g.fpsObject.Update()
		now := time.Now()
		g.handleKeyInput(dt)
		g.players.Range(func(k, v interface{}) bool {
			p := v.(*world.Player)
			p.Update(dt)
			return true
		})
		dur := now.Sub(prev)
		dt = float64(dur) / float64(time.Second)
		if dur < delta {
			time.Sleep(delta - dur)
		}
		prev = now

	}
}

func (g *Game) Update() {
	var dt float64
	now := glfw.GetTime()
	dt = now - g.prevtime
	g.prevtime = now
	if dt > 0.02 {
		dt = 0.02
	}
	//g.player.Update(dt)
	mainthread.Call(func() {
		g.fps.Update()
		gl.ClearColor(0.57, 0.71, 0.77, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		g.blockRender.Draw(g.player)
		g.lineRender.Draw(g.player)
		mat := g.blockRender.Get3dmat(g.player)
		//g.playerRender.DrawPlayer(g.player, g.player)
		g.playerRender.Draw(mat, g.players)
		g.renderStat()

		g.win.SwapBuffers()
		glfw.PollEvents()
		g.closed = g.win.ShouldClose()
	})
}
