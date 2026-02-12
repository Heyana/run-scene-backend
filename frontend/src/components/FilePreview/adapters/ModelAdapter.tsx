import { defineComponent, ref, onMounted, onUnmounted } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

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

    // TODO: 初始化 3D 渲染器（Three.js）
    const initRenderer = () => {
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
        setTimeout(() => {
          loading.value = false;
          props.onLoad?.();
        }, 1000);
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
            <div
              style={{
                textAlign: "center",
                paddingTop: "200px",
                color: "#999",
              }}
            >
              TODO: Three.js 3D 模型渲染
            </div>
          </div>
        )}
      </div>
    );
  },
});

export default new ModelPreviewAdapter();
