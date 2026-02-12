import { defineComponent, ref, onMounted, onUnmounted, reactive } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";
import type RunScene from "run-scene-v2/types/src/RunScene";
import ThreeScene from "@/components/three-scene/index.vue";
import "./styles/model.less";
// 3D 模型预览适配器
class ModelPreviewAdapter implements IPreviewAdapter {
  name = "ModelPreviewAdapter";

  private supportedFormats = ["glb", "gltf", "fbx", "obj"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <ModelPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 3D 模型预览组件
const ModelPreview = defineComponent({
  name: "ModelPreview",
  props: {
    file: {
      type: Object,
      required: true,
    },
    onLoad: Function,
    onError: Function,
  },
  setup(props) {
    const containerRef = ref<HTMLDivElement>();
    const loading = ref(true);
    const error = ref(false);
    const url =
      "http://192.168.3.8:8080/file?path=project/linkpoint/&key=202602121521042811001001202673";
    const defRunSceneConfig = {
      renderConfig: {
        matrixAutoUpdate: true,
        scriptFrame: 60,
        event: {
          // ignores: ["resize"],
        },
      },
      // showFps: getEnvMode() === "local",

      camera: {
        showBackground: true,
      },
    };
    const constOvewview = {
      options: {
        ...defRunSceneConfig,
        ltPp: {
          modules: {
            ignores: ["SelectiveBloom", "Outline", "Outline1"],
          },
        },
        // mode: "editor",

        renderConfig: {
          // matrixAutoUpdate: true,
          scriptFrame: 60,
          event: {
            // ignores: ["resize"],
          },
          // frame: 30,

          getSize: () => {
            const dom = document.querySelector(".model-canvas-container");
            const b = dom?.getBoundingClientRect();
            console.log("Log-- ", b, "b");
            return {
              width: 1200,
              height: document.body.getBoundingClientRect().height * 0.8,
            };
          },
        },
        loadConfig: {
          // lazy: true,
          block: {
            paths: [],
          },
          engineDom: {
            forceFullSize: true,
          },
        },
      },
    };

    // 根据文件格式获取 MIME 类型
    const getMimeType = (format: string): string => {
      const mimeTypes: Record<string, string> = {
        glb: "model/gltf-binary",
        gltf: "model/gltf+json",
        fbx: "application/octet-stream",
        obj: "text/plain",
      };
      return mimeTypes[format.toLowerCase()] || "application/octet-stream";
    };

    // 下载文件并转换为 File 对象
    const downloadFile = async () => {
      try {
        const response = await fetch(props.file.file_url);
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const blob = await response.blob();
        const mimeType = getMimeType(props.file.format);
        const file = new File(
          [blob],
          props.file.name + "." + props.file.format,
          {
            type: mimeType,
          },
        );

        console.log("文件下载完成:", file, "MIME类型:", mimeType);
        return file;
      } catch (err) {
        console.error("文件下载失败:", err);
        error.value = true;
        props.onError?.(err as Error);
      }
    };
    let runScene: RunScene | undefined;
    const pageScene = new (class {
      options = constOvewview.options;
      onPreLoaded = async (theRunScene: RunScene) => {
        // const ls = [1, 2, 3, 4, 5];
        // ls.reverse().map((i) => {
        //   setTimeout(async () => {
        //     await theRunScene.cameraEx.setTemp(i + "", {
        //       time: 0.1,
        //       onComplete: () => {
        //         console.log("Log-- ", "onComplete");
        //       },
        //     });
        //   }, 1);
        // });

        console.log("Log-- ", theRunScene, "theRunScene");
      };
      getPath() {
        return url;
      }
      data = reactive({});
      onLoaded = async (
        theRunScene: RunScene,
        map: {
          dom: HTMLElement;
        },
      ) => {
        runScene = theRunScene;
        // 先下载文件
        const file = await downloadFile();

        if (!file) {
          throw new Error("文件下载失败");
        }
        const results = await theRunScene.fileEx.parseFiles([file], {
          clearMaterial: false,
        });
        console.log("Log-- ", results, "results");
        results.map((map) => {
          const { result, type, file } = map;
          console.log("Log-- ", map, "map");
          if (type === "model") {
            theRunScene.modelEx.add(result, undefined, {
              isClone: true,
              select: true,
            });
            theRunScene.modelEx.focus(result[0]);

            theRunScene.cb.loaderer.gltf.modelAdded.cb({ models: result });
          }
        });
      };
    })();
    // TODO: 初始化 3D 渲染器（Three.js）
    const initRenderer = async () => {
      try {
        loading.value = true;
        error.value = false;

        // 插槽：在这里实现 Three.js 场景初始化
        // 1. 创建 Scene, Camera, Renderer
        // 2. 添加光源
        // 3. 加载模型（根据 props.file.format 选择加载器：GLTFLoader, FBXLoader, OBJLoader）
        // 4. 添加 OrbitControls
        // 5. 启动渲染循环

        console.log("TODO: 初始化 3D 渲染器", props.file);

        // 模拟加载完成
        loading.value = false;
        props.onLoad?.();
      } catch (err: any) {
        error.value = true;
        loading.value = false;
        props.onError?.(err);
      }
    };

    // TODO: 清理资源
    const cleanup = () => {
      // 插槽：在这里清理 Three.js 资源
      // 1. 停止渲染循环
      // 2. 释放几何体、材质、纹理
      // 3. 销毁渲染器
      runScene?.clean();
      console.log("TODO: 清理 3D 渲染器资源");
    };

    onMounted(() => {
      initRenderer();
    });

    onUnmounted(() => {
      cleanup();
    });

    return () => (
      <div class="model-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载模型中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">🎨</div>
            <div class="error-text">模型加载失败</div>
            <div class="error-hint">请检查文件格式或网络连接</div>
          </div>
        ) : (
          <div
            ref={containerRef}
            class="model-canvas-container"
            style={{
              width: "100%",
              height: "80vh",
              display: loading.value ? "none" : "block",
            }}
          >
            {/* Three.js 渲染器将挂载到这里 */}

            <ThreeScene
              key="overview-three-scene"
              class="three-scene"
              ref="childComp"
              type="scene"
              options={pageScene.options}
              onLoaded={pageScene.onLoaded}
              onPreLoaded={pageScene.onPreLoaded}
              path={pageScene.getPath()}
            ></ThreeScene>
          </div>
        )}
      </div>
    );
  },
});

export default new ModelPreviewAdapter();
