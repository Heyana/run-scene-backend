import { defineComponent, ref, onMounted } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// HTML 预览适配器
class HtmlPreviewAdapter implements IPreviewAdapter {
  name = "HtmlPreviewAdapter";

  private supportedFormats = ["html", "htm"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <HtmlPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// HTML 预览组件
const HtmlPreview = defineComponent({
  name: "HtmlPreview",
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
    const iframeRef = ref<HTMLIFrameElement>();

    const handleLoad = () => {
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = () => {
      loading.value = false;
      error.value = true;
      props.onError?.(new Error("HTML 加载失败"));
    };

    onMounted(() => {
      if (iframeRef.value) {
        iframeRef.value.addEventListener("load", handleLoad);
        iframeRef.value.addEventListener("error", handleError);
      }
    });

    return () => (
      <div class="html-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载 HTML 中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">🌐</div>
            <div class="error-text">HTML 加载失败</div>
            <div class="error-hint">请检查文件或网络连接</div>
          </div>
        ) : (
          <iframe
            ref={iframeRef}
            src={props.file.file_url}
            style={{
              width: "100%",
              height: "80vh",
              border: "none",
              display: loading.value ? "none" : "block",
              backgroundColor: "#fff",
            }}
            title={props.file.name}
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-modals"
          />
        )}
      </div>
    );
  },
});

export default new HtmlPreviewAdapter();
