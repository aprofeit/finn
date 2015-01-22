KeyboardEvents =
  bind: (websocket) ->
    window.onkeydown = (e) ->
      websocket.send(JSON.stringify(event: "keydown", keycode: e.keyCode))

    window.onkeyup = (e) ->
      websocket.send(JSON.stringify(event: "keyup", keycode: e.keyCode))

Tiles = Backbone.Collection.extend
  url: ->
    "/tiles"

Player = Backbone.Model.extend
  setAnimationFrame: ->
    walkFrames = [1, 3]
    frame = (parseInt(Date.now() / 200) % 2)
    filename = switch @get("direction")
      when "up"
        @lastDirection = "north"
        "north#{walkFrames[frame]}"
      when "down"
        @lastDirection = "south"
        "south#{walkFrames[frame]}"
      when "left"
        @lastDirection = "west"
        "west#{walkFrames[frame]}"
      when "right"
        @lastDirection = "east"
        "east#{walkFrames[frame]}"
      when "none"
        if @lastDirection
          "#{@lastDirection}2"
        else
          "south2"

    @sprite.setTexture(PIXI.Texture.fromImage("sprites/#{filename}.png"))

Players = Backbone.Collection.extend
  model: Player

class World
  sort: ->
    @stage.children.sort (a, b) ->
      if (a.z < b.z)
        return -1
      else if (a.z > b.z)
        return 1
      else
        return 0

  constructor: (options) ->
    @stage = new PIXI.Stage(0x303030)
    @renderer = PIXI.autoDetectRenderer(1000, 600)
    @projectiles = new Backbone.Collection()
    @members = new Players()
    @currentPlayer = new Player()
    @addedPlayerToStage = false
    @tiles = new Tiles()
    document.body.appendChild(@renderer.view)

    @currentPlayer.on "change", (player) =>
      if !@addedPlayerToStage
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage(player.get("texture")))
        sprite.z = 1
        sprite.position.x = 500
        sprite.position.y = 300
        sprite.height = player.get("height") * 100
        sprite.width = player.get("width") * 100
        @stage.addChild(sprite)
        player.sprite = sprite
        @sort()
        @addedPlayerToStage = true

      if player.get("dead")
        @stage.removeChild(player.sprite)
      else
        @xOff = ((1000/100)/2.0) - player.get("position_x")
        @yOff = ((600/100)/2.0) - player.get("position_y")

        @members.forEach (player) =>
          player.sprite.position.x = (player.get("position_x") + @xOff) * 100
          player.sprite.position.y = (player.get("position_y") + @yOff) * 100

        @tiles.forEach (tile) =>
          tile.sprite.position.x = (tile.get("x") + @xOff) * 100
          tile.sprite.position.y = (tile.get("y") + @yOff) * 100

        @projectiles.forEach (projectile) =>
          projectile.sprite.position.x = (projectile.get("x") + @xOff) * 100
          projectile.sprite.position.y = (projectile.get("y") + @yOff) * 100

    @tiles.on "sync", =>
      @sort()

    @tiles.on "add", (tile) =>
      if tile.get("kind") == "wall"
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage("sprites/wall.png"))
        sprite.z = 0
        sprite.position.x = (tile.get("x") + @xOff) * 100
        sprite.position.y = (tile.get("y") + @yOff) * 100
        sprite.width = 100
        sprite.height = 100
        @stage.addChild(sprite)
        tile.sprite = sprite
      else if tile.get("kind") == "floor"
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage("sprites/grass.png"))
        sprite.z = 0
        sprite.position.x = (tile.get("x") + @xOff) * 100
        sprite.position.y = (tile.get("y") + @yOff) * 100
        sprite.width = 100
        sprite.height = 100
        @stage.addChild(sprite)
        tile.sprite = sprite

    @projectiles.on "add", (projectile) =>
      sprite = new PIXI.Sprite(PIXI.Texture.fromImage(projectile.get("texture")))
      sprite.z = 0.1
      sprite.position.x = (projectile.get("position_x") + @xOff) * 100
      sprite.position.y = (projectile.get("position_y") + @yOff)* 100
      @stage.addChild(sprite)
      projectile.sprite = sprite

    @projectiles.on "change", (projectile) =>
      sprite = projectile.sprite
      sprite.position.x = (projectile.get("position_x") + @xOff) * 100
      sprite.position.y = (projectile.get("position_y") + @yOff)* 100

    @projectiles.on "remove", (projectile) =>
      @stage.removeChild(projectile.sprite)

    @members.on "add", (player) =>
      sprite = new PIXI.Sprite(PIXI.Texture.fromImage(player.get("texture")))
      sprite.z = 1
      sprite.position.x = (player.get("position_x") + @xOff) * 100
      sprite.position.y = (player.get("position_y") + @yOff)* 100
      sprite.height = player.get("height") * 100
      sprite.width = player.get("width") * 100
      @stage.addChild(sprite)
      player.sprite = sprite
      @sort()

    @members.on "remove", (player) =>
      @stage.removeChild(player.sprite)

    @members.on "change", (player) =>
      if player.get("dead")
        @stage.removeChild(player.sprite)
      else
        sprite = player.sprite
        sprite.position.x = (player.get("position_x") + @xOff) * 100
        sprite.position.y = (player.get("position_y") + @yOff)* 100

  connect: ->
    @websocket = new WebSocket("ws://#{window.location.host}/websocket")
    @websocket.onmessage = (e) =>
      world = JSON.parse(e.data)
      @update(world)

    @websocket.onopen = =>
      @tiles.fetch()

  update: (update) ->
    @currentPlayer.set(update.current)
    @members.set(update.members)
    @projectiles.set(update.projectiles)
    @members.forEach (player) ->
      player.setAnimationFrame()
    @currentPlayer.setAnimationFrame()

  render: (elapsed) =>
    requestAnimFrame(@render)
    @renderer.render(@stage)

world = new World()
world.connect()
world.render()

KeyboardEvents.bind(world.websocket)
