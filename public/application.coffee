KeyboardEvents =
  bind: (websocket) ->
    window.onkeydown = (e) ->
      websocket.send(JSON.stringify(event: "keydown", keycode: e.keyCode))

    window.onkeyup = (e) ->
      console.log e
      websocket.send(JSON.stringify(event: "keyup", keycode: e.keyCode))

class World
  constructor: (options) ->
    @stage = new PIXI.Stage(0xFFFFFF)
    @renderer = PIXI.autoDetectRenderer(1000, 600)
    @members = new Backbone.Collection()
    document.body.appendChild(@renderer.view)

    @members.on "add", (player) =>
      sprite = new PIXI.Sprite(PIXI.Texture.fromImage(player.get("texture")))
      sprite.anchor.x = player.get("anchor_x")
      sprite.anchor.y = player.get("anchor_y")
      sprite.position.x = player.get("position_x")
      sprite.position.y = player.get("position_y")
      @stage.addChild(sprite)
      player.sprite = sprite

    @members.on "remove", (player) =>
      @stage.removeChild(player.sprite)

    @members.on "change", (player) ->
      sprite = player.sprite
      sprite.position.x = player.get("position_x")
      sprite.position.y = player.get("position_y")

  connect: ->
    @websocket = new WebSocket("ws://#{window.location.host}/websocket")
    @websocket.onmessage = (e) =>
      world = JSON.parse(e.data)
      @update(world)

  update: (update) ->
    @members.set(update.members)
    @members.forEach (player) ->
      walkFrames = [1, 3]
      frame = (parseInt(Date.now() / 200) % 2)
      filename = switch player.get("direction")
        when "up"
          player.lastDirection = "north"
          "north#{walkFrames[frame]}"
        when "down"
          player.lastDirection = "south"
          "south#{walkFrames[frame]}"
        when "left"
          player.lastDirection = "west"
          "west#{walkFrames[frame]}"
        when "right"
          player.lastDirection = "east"
          "east#{walkFrames[frame]}"
        when "none"
          if player.lastDirection
            "#{player.lastDirection}2"
          else
            "south2"

      player.sprite.setTexture(PIXI.Texture.fromImage("sprites/#{filename}.png"))

  render: (elapsed) =>
    requestAnimFrame(@render)
    @renderer.render(@stage)

world = new World()
world.connect()
world.render()

KeyboardEvents.bind(world.websocket)
