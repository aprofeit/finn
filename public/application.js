// Generated by CoffeeScript 1.4.0
var KeyboardEvents, Player, Players, ScoreView, Tiles, World, scoreView, world,
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

Player = Backbone.Model.extend({
  setAnimationFrame: function() {
    var filename, frame, walkFrames;
    walkFrames = [1, 3];
    frame = parseInt(Date.now() / 200) % 2;
    filename = (function() {
      switch (this.get("direction")) {
        case "up":
          this.lastDirection = "north";
          return "north" + walkFrames[frame];
        case "down":
          this.lastDirection = "south";
          return "south" + walkFrames[frame];
        case "left":
          this.lastDirection = "west";
          return "west" + walkFrames[frame];
        case "right":
          this.lastDirection = "east";
          return "east" + walkFrames[frame];
        case "none":
          if (this.lastDirection) {
            return "" + this.lastDirection + "2";
          } else {
            return "south2";
          }
      }
    }).call(this);
    return this.sprite.setTexture(PIXI.Texture.fromImage("sprites/" + filename + ".png"));
  }
});

Players = Backbone.Collection.extend({
  model: Player
});

ScoreView = Backbone.View.extend({
  tagName: "div",
  className: "score-container",
  initialize: function() {
    this.$el.append($("<div class='score'></div>"));
    this.$el.append($("<div class='high-score'></div>"));
    return this.listenTo(this.model, "change", this.render);
  },
  render: function() {
    console.log(this.$el);
    this.$(".score").html("SCORE " + (this.model.get("score")));
    this.$(".high-score").html("HIGH " + (this.model.get("high_score")));
    return this.$el;
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
    this.projectiles = new Backbone.Collection();
    this.members = new Players();
    this.currentPlayer = new Player();
    this.addedPlayerToStage = false;
    this.tiles = new Tiles();
    document.body.appendChild(this.renderer.view);
    this.currentPlayer.on("change", function(player) {
      var sprite;
      if (!_this.addedPlayerToStage) {
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage(player.get("texture")));
        sprite.z = 1;
        sprite.position.x = 500;
        sprite.position.y = 300;
        sprite.height = player.get("height") * 100;
        sprite.width = player.get("width") * 100;
        _this.stage.addChild(sprite);
        player.sprite = sprite;
        _this.sort();
        _this.addedPlayerToStage = true;
      }
      if (player.get("dead")) {
        return _this.stage.removeChild(player.sprite);
      } else {
        _this.xOff = ((1000 / 100) / 2.0) - player.get("position_x");
        _this.yOff = ((600 / 100) / 2.0) - player.get("position_y");
        _this.members.forEach(function(player) {
          player.sprite.position.x = (player.get("position_x") + _this.xOff) * 100;
          return player.sprite.position.y = (player.get("position_y") + _this.yOff) * 100;
        });
        _this.tiles.forEach(function(tile) {
          tile.sprite.position.x = (tile.get("x") + _this.xOff) * 100;
          return tile.sprite.position.y = (tile.get("y") + _this.yOff) * 100;
        });
        return _this.projectiles.forEach(function(projectile) {
          projectile.sprite.position.x = (projectile.get("x") + _this.xOff) * 100;
          return projectile.sprite.position.y = (projectile.get("y") + _this.yOff) * 100;
        });
      }
    });
    this.tiles.on("sync", function() {
      return _this.sort();
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
        return tile.sprite = sprite;
      } else if (tile.get("kind") === "floor") {
        sprite = new PIXI.Sprite(PIXI.Texture.fromImage("sprites/grass.png"));
        sprite.z = 0;
        sprite.position.x = (tile.get("x") + _this.xOff) * 100;
        sprite.position.y = (tile.get("y") + _this.yOff) * 100;
        sprite.width = 100;
        sprite.height = 100;
        _this.stage.addChild(sprite);
        return tile.sprite = sprite;
      }
    });
    this.projectiles.on("add", function(projectile) {
      var sprite;
      sprite = new PIXI.Sprite(PIXI.Texture.fromImage(projectile.get("texture")));
      sprite.z = 0.1;
      sprite.position.x = (projectile.get("position_x") + _this.xOff) * 100;
      sprite.position.y = (projectile.get("position_y") + _this.yOff) * 100;
      _this.stage.addChild(sprite);
      return projectile.sprite = sprite;
    });
    this.projectiles.on("change", function(projectile) {
      var sprite;
      sprite = projectile.sprite;
      sprite.position.x = (projectile.get("position_x") + _this.xOff) * 100;
      return sprite.position.y = (projectile.get("position_y") + _this.yOff) * 100;
    });
    this.projectiles.on("remove", function(projectile) {
      return _this.stage.removeChild(projectile.sprite);
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
    this.members.on("change", function(player, options) {
      var sprite;
      if (player.get("dead")) {
        return _this.stage.removeChild(player.sprite);
      } else {
        sprite = player.sprite;
        sprite.position.x = (player.get("position_x") + _this.xOff) * 100;
        return sprite.position.y = (player.get("position_y") + _this.yOff) * 100;
      }
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
    this.projectiles.set(update.projectiles);
    this.members.forEach(function(player) {
      return player.setAnimationFrame();
    });
    return this.currentPlayer.setAnimationFrame();
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

scoreView = new ScoreView({
  model: world.currentPlayer
}).render();

$("body").append(scoreView);

KeyboardEvents.bind(world.websocket);
