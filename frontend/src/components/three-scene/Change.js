const fn = (runScene, inputData = {}, constant = {}) => {
  const fn = (map) => {
    const {
      runScene,
      // Utils,
      // core,
      // getModel,
      // constant,
      bus,
      // Three,
      // camera,
      // scene,
      controls,
      // renderer,
    } = map;

    // 场景初始化
    class InitScene {
      name = "initScene";
      mounted() {}
    }

    // 基本事件
    class Events {
      name = "events";
      constructor() {
        runScene.cb.model.setSelect.add(
          "trigger-click",
          this.triggerClick.bind(this)
        );
      }
      triggerClick = (model) => {
        if (!model) return;
        bus.$emit("logClickModel", model);
      };

      dispose() {
        controls.removeEventListener("start", this.controlStart);
      }
    }

    return [Events, InitScene];
  };

  const modules = fn({
    runScene,
    getModel: runScene.modelEx.getModel.bind(runScene.modelEx),
    core: runScene.custom,
    ...runScene.assetsEx.get(),
    ...inputData,
    constant,
    window: null,
  });

  if (!modules) return;

  modules
    .map((TheClass) => {
      const ins = new TheClass();
      if (!ins.name) throw TypeError("代码出错");
      runScene.custom[ins.name] = ins;
      return ins;
    })
    .map((ins) => ins?.mounted?.());
};

export { fn };
