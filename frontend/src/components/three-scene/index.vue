<template>
  <div>
    <div
      class="scene"
      :style="{
        backgroundImage: `url(  ./images/scene_bgc1.png)`,
      }"
      ref="scene"
      onselectstart="return false;"
    ></div>
    <div v-show="!loaded" class="load">
      <div class="mask">
        <div class="inner">
          <div
            class="tip"
            :style="{
              backgroundImage: `url(./images/提示.png)`,
            }"
          >
            <img :src="`./images/提示1.png`" alt="" />
          </div>

          <div class="loading">
            <div class="loading_inner" :style="{ width: progress + '%' }"></div>
          </div>
        </div>
      </div>
      <!-- <div v-show="!loaded" class="load"> -->
    </div>
  </div>
</template>
<script>
// 场景
const runSceneMap = {};
import * as Engine from "run-scene-v2/build/index";
import { fn } from "./Change";
import bus from "./Bus";
import { nextTick } from "vue";

const { RunScene, Utils, Three } = Engine;
export default {
  props: {
    path: String,
    type: String,
    onLoaded: Function,
    onPreLoaded: Function,
    backgroundPath: String,
    options: Object,
  },
  name: "ThreeScene",
  data() {
    return {
      loaded: false,
      progress: 0, // 加载进度 0-100
    };
  },
  async mounted(props) {
    nextTick(() => {
      this.load();
    });
  },
  methods: {
    async load() {
      const { path, type, backgroundPath } = this.$props;
      const runScene = await this.loadScene(path, this.$props);
      console.log("Log-- ", this.$props, "this.$props");
      console.log("Log-- ", runScene, this.$props, "runScene");
      this.setBgc(backgroundPath, runScene);

      runSceneMap[type] = runScene;

      runSceneMap[type].on("loaded", async () => {
        const { type, onLoaded, onPreLoaded } = this.$props;
        if (onPreLoaded) {
          await onPreLoaded(runSceneMap[type]);
        }
        runScene.graph.run();
        runScene.sceneEx.autoOpenUpdate();
        runScene.script.playAll();
        // 自适应
        // loading
        this.loaded = true;

        // 场景功能
        fn(runScene, {
          Utils,
          bus,
          Three,
        });
      });
    },

    // 场景加载
    loadScene(path, props) {
      const { type, onLoaded, onPreLoaded } = this.$props;

      console.log("Log-- ", type, props, "type");
      const runScene = new RunScene(props.options || defRunSceneConfig)
        .load({
          path: path,
          dom: this.$refs["scene"],
        })
        .on("getJSON", () => {
          console.time("开始加载场景");
        })
        .on("loaded", async () => {
          console.log("Log-- ", path, onPreLoaded, "path");
          // runScene.renderEx.setSize(1920, 1080, true);

          this.progress = 100; // 加载完成

          onLoaded(runSceneMap[type], {
            dom: this.$refs["scene"],
          });
        })
        .on("progress", (progress) => {
          console.log("Log-- ", progress, "progress");
          this.progress = progress; // 更新进度
        });

      return runScene;
    },

    resize: (x, runScene) => {
      // const map = runScene.assetsEx.engineDom.getBoundingClientRect();
      // runScene.renderEx.setSize(map.width / x, map.height / x);
    },

    async setBgc(path, runScene) {
      if (!path) return;
      const dom = runScene.assetsEx.engineDom;
      const backgroudDom = document.createElement("div");
      backgroudDom.style.position = "absolute";
      backgroudDom.style.left = "50%";
      backgroudDom.style.top = "50%";
      backgroudDom.style.transform = "translate(-50%,-50%)";
      backgroudDom.style.width = "100%";
      backgroudDom.style.height = "100%";
      dom.appendChild(backgroudDom);
      backgroudDom.style.backgroundImage = "url(" + path + ")";
      backgroudDom.style.backgroundRepeat = "no-repeat";
      backgroudDom.style.backgroundSize = "100% 100%";
    },
  },
};
</script>

<style lang="less" scoped>
.mask {
  width: 100%;
  height: 100%;
  position: absolute;
  z-index: 50;
  background-size: 100% 100%;

  .inner {
    position: absolute;
    width: 90px;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    z-index: 10;

    img {
      .full;
      object-fit: contain;
    }

    .tip {
      .w;
      height: 57px;
    }

    .loading {
      .w;
      margin-top: 10px;
      height: 7px;
      border-radius: 90px;
      background: rgba(255, 255, 255, 0.3);

      &_inner {
        // width: 50%;
        .h;
        border-radius: 90px;
        background: linear-gradient(90deg, #007aff 0%, #00b2ff 100%), #5891ff;
      }
    }
  }
}

.load {
  position: absolute;
  z-index: 50;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10.449999809265137px);
}

// 场景
.scene {
  // width: 100%;
  // height: 100%;
  position: absolute;
  z-index: 0;
  left: 0;
  top: 0;
}

.text {
  width: 250px;
}

.scene .btn {
  position: absolute;
  z-index: 2;
}

.scene .show {
  opacity: 1 !important;
}

.scene .none {
  opacity: 0 !important;
}

.scene .block {
  display: block !important;
}

/* 新的loading动画 */
.spinner {
  position: relative;
  width: 80px;
  height: 30px;
}

.dot1,
.dot2,
.dot3 {
  position: absolute;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background-color: #00b4d8;
  /* 亮青色 */
  top: 0;
  animation: bounce 1.4s infinite ease-in-out both;
}

.dot1 {
  left: 0;
  animation-delay: -0.32s;
}

.dot2 {
  left: 32px;
  animation-delay: -0.16s;
  background-color: #90e0ef;
  /* 浅青色 */
}

.dot3 {
  left: 64px;
  background-color: #48cae4;
  /* 中青色 */
}

@keyframes bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
  }

  40% {
    transform: translateY(-20px);
  }
}
</style>
