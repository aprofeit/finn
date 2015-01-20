// Generated by CoffeeScript 1.4.0
var KeyboardEvents, Tiles, World, world,
  __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

KeyboardEvents = {
  bind: function(websocket) {
    window.onkeydown = function(e) {
      return websocket.send(JSON.stringify({
        event: "keydown",
        keycode: e.keyCode
      }));
    };
    return window.onkeyup = function(e) {
      return websocket.send(JSON.stringify({
        event: "keyup",
        keycode: e.keyCode
      }));
    };
  }
};

Tiles = Backbone.Collection.extend({
  url: function() {
    return "/tiles";
  }
});

World = (function() {

  World.prototype.sort = function() {
    return this.stage.children.sort(function(a, b) {
      if (a.z < b.z) {
        return -1;
      } else if (a.z > b.z) {
        return 1;
      } else {
        return 0;
      }
    });
  };

  function World(options) {
    this.render = __bind(this.render, this);

    var _this = this;
    this.stage = new PIXI.Stage(0x303030);
    this.renderer = PIXI.autoDetectRenderer(1000, 600);
    this.members = new Backbone.Collection();
    this.currentPlayer = new Backbone.Model();
    this.tiles = new Tiles();
    document.body.appendChild(this.renderer.view);
    this.currentPlayer.on("change", function(player) {
      _this.xOff = ((1000 / 100) / 2.0) - player.get("position_x");
      _this.yOff = ((600 / 100) / 2.0) - player.get("position_y");
      return _this.tiles.forEach(function(tile) {
        tile.sprite.position.x = (tile.get("x") + _this.xOff) * 100;
        return tile.sprite.position.y = (tile.get("y") + _this.yOff) * 100;
      });
    });
    this.tiles.on("add", function(tile) {
      var sprite;
      if (tile.get("kind") === "wall") {
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage("sprites/wall.png"));
        sprite.z = 0;
        sprite.position.x = (tile.get("x") + _this.xOff) * 100;
        sprite.position.y = (tile.get("y") + _this.yOff) * 100;
        sprite.width = 100;
        sprite.height = 100;
        _this.stage.addChild(sprite);
        tile.sprite = sprite;
      } else if (tile.get("kind") === "floor") {
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage("sprites/grass.png"));
        sprite.z = 0;
        sprite.position.x = (tile.get("x") + _this.xOff) * 100;
        sprite.position.y = (tile.get("y") + _this.yOff) * 100;
        sprite.width = 100;
        sprite.height = 100;
        _this.stage.addChild(sprite);
        tile.sprite = sprite;
      }
      return _this.sort();
    });
    this.members.on("add", function(player) {
      var sprite;
      sprite = new PIXI.Sprite(PIXI.Texture.fromImage(player.get("texture")));
      sprite.z = 1;
      sprite.position.x = (player.get("position_x") + _this.xOff) * 100;
      sprite.position.y = (player.get("position_y") + _this.yOff) * 100;
      sprite.height = player.get("height") * 100;
      sprite.width = player.get("width") * 100;
      _this.stage.addChild(sprite);
      player.sprite = sprite;
      return _this.sort();
    });
    this.members.on("remove", function(player) {
      return _this.stage.removeChild(player.sprite);
    });
    this.members.on("change", function(player) {
      var sprite;
      sprite = player.sprite;
      sprite.position.x = 500;
      return sprite.position.y = 300;
    });
  }

  World.prototype.connect = function() {
    var _this = this;
    this.websocket = new WebSocket("ws://" + window.location.host + "/websocket");
    this.websocket.onmessage = function(e) {
      var world;
      world = JSON.parse(e.data);
      return _this.update(world);
    };
    return this.websocket.onopen = function() {
      return _this.tiles.fetch();
    };
  };

  World.prototype.update = function(update) {
    this.currentPlayer.set(update.current);
    this.members.set(update.members);
    return this.members.forEach(function(player) {
      var filename, frame, walkFrames;
      walkFrames = [1, 3];
      frame = parseInt(Date.now() / 200) % 2;
      filename = (function() {
        switch (player.get("direction")) {
          case "up":
            player.lastDirection = "north";
            return "north" + walkFrames[frame];
          case "down":
            player.lastDirection = "south";
            return "south" + walkFrames[frame];
          case "left":
            player.lastDirection = "west";
            return "west" + walkFrames[frame];
          case "right":
            player.lastDirection = "east";
            return "east" + walkFrames[frame];
          case "none":
            if (player.lastDirection) {
              return "" + player.lastDirection + "2";
            } else {
              return "south2";
            }
        }
      })();
      return player.sprite.setTexture(PIXI.Texture.fromImage("sprites/" + filename + ".png"));
    });
  };

  World.prototype.render = function(elapsed) {
    requestAnimFrame(this.render);
    return this.renderer.render(this.stage);
  };

  return World;

})();

world = new World();

world.connect();

world.render();

KeyboardEvents.bind(world.websocket);
