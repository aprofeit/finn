KeyboardEvents =
  bind: (websocket) ->
    window.onkeydown = (e) ->
      websocket.send(JSON.stringify(event: "keydown", keycode: e.keyCode))

    window.onkeyup = (e) ->
      websocket.send(JSON.stringify(event: "keyup", keycode: e.keyCode))

class World
  constructor: (options) ->
    @stage = new PIXI.Stage(0xFFFFFF)
    @renderer = PIXI.autoDetectRenderer(1000, 600)
    @members = {}
    document.body.appendChild(@renderer.view)

  connect: ->
    @websocket = new WebSocket("ws://#{window.location.host}/websocket")
    @websocket.onmessage = (e) =>
      world = JSON.parse(e.data)
      @update(world)

  update: (update) ->
    ids = _.map update.members, (m) ->
      m.id

    for id, member of @members
      if ids.indexOf(id) == -1
        @stage.removeChild(member.sprite)
        delete @members[id]

    for member in update.members
      if @members[member.id]
        sprite = @members[member.id].sprite
        sprite.position.x = member.position_x
        sprite.position.y = member.position_y
      else
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage(member.texture))
        sprite.anchor.x = member.anchor_x
        sprite.anchor.y = member.anchor_y
        sprite.position.x = member.position_x
        sprite.position.y = member.position_y
        @stage.addChild(sprite)
        member.sprite = sprite
        @members[member.id] = member

  render: (elapsed) =>
    requestAnimFrame(@render)
    @renderer.render(@stage)

world = new World()
world.connect()
world.render()

KeyboardEvents.bind(world.websocket)
