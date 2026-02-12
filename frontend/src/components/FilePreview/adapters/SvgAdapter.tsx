import { defineComponent, ref, onMounted } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";
import "./styles/svg.less";

// SVG 预览适配器
class SvgPreviewAdapter implements IPreviewAdapter {
  name = "SvgPreviewAdapter";

  canPreview(format: string): boolean {
    return format.toLowerCase() === "svg";
  }

  render(props: PreviewAdapterProps) {
    return (
      <SvgPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// SVG 预览组件
const SvgPreview = defineComponent({
  name: "SvgPreview",
  props: {
    file: {
      type: Object,
      required: true,
    },
    onLoad: Function,
    onError: Function,
  },
  setup(props) {
    const loading = ref(true);
    const error = ref(false);
    const svgContent = ref("");
    const scale = ref(1);
    const containerRef = ref<HTMLDivElement>();

    // 加载 SVG 内容
    const loadSvg = async () => {
      try {
        loading.value = true;
        error.value = false;

        const response = await fetch(props.file.file_url);
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const text = await response.text();
        svgContent.value = text;
        loading.value = false;
        props.onLoad?.();
      } catch (err: any) {
        console.error("SVG 加载失败:", err);
        error.value = true;
        loading.value = false;
        props.onError?.(err);
      }
    };

    // 缩放控制
    const handleZoomIn = () => {
      scale.value = Math.min(scale.value + 0.2, 5);
    };

    const handleZoomOut = () => {
      scale.value = Math.max(scale.value - 0.2, 0.2);
    };

    const handleResetZoom = () => {
      scale.value = 1;
    };

    // 在新窗口打开
    const handleOpenInNewWindow = () => {
      window.open(props.file.file_url, "_blank");
    };

    onMounted(() => {
      loadSvg();
    });

    return () => (
      <div class="svg-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载 SVG 中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">🖼️</div>
            <div class="error-text">SVG 加载失败</div>
            <div class="error-hint">请检查文件格式或网络连接</div>
          </div>
        ) : (
          !loading.value && (
            <>
              {/* 工具栏 */}
              <div class="svg-toolbar">
                <button onClick={handleZoomIn} title="放大">
                  +
                </button>
                <button onClick={handleZoomOut} title="缩小">
                  -
                </button>
                <button onClick={handleResetZoom} title="重置">
                  {Math.round(scale.value * 100)}%
                </button>
                <button onClick={handleOpenInNewWindow} title="新窗口打开">
                  ↗
                </button>
              </div>

              {/* SVG 内容 */}
              <div ref={containerRef} class="svg-content-wrapper">
                <div
                  class="svg-content"
                  style={{
                    transform: `scale(${scale.value})`,
                  }}
                  innerHTML={svgContent.value}
                />
              </div>
            </>
          )
        )}
      </div>
    );
  },
});

export default new SvgPreviewAdapter();
