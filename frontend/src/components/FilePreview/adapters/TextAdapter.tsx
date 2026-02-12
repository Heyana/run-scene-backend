import { defineComponent, ref, onMounted } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// 文本预览适配器
class TextPreviewAdapter implements IPreviewAdapter {
  name = "TextPreviewAdapter";

  private supportedFormats = ["txt", "md", "log", "json", "xml", "csv"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <TextPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 文本预览组件
const TextPreview = defineComponent({
  name: "TextPreview",
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
    const content = ref<string>("");

    const loadText = async () => {
      try {
        loading.value = true;
        error.value = false;

        const response = await fetch(props.file.file_url);
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const text = await response.text();
        content.value = text;
        loading.value = false;
        props.onLoad?.();
      } catch (err: any) {
        error.value = true;
        loading.value = false;
        props.onError?.(err);
      }
    };

    onMounted(() => {
      loadText();
    });

    return () => (
      <div class="text-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">📄</div>
            <div class="error-text">文本加载失败</div>
            <div class="error-hint">请检查文件或网络连接</div>
          </div>
        ) : (
          <div
            class="text-content"
            style={{ display: loading.value ? "none" : "block" }}
          >
            <pre
              style={{
                padding: "20px",
                backgroundColor: "#f5f5f5",
                borderRadius: "4px",
                maxHeight: "80vh",
                overflow: "auto",
                whiteSpace: "pre-wrap",
                wordWrap: "break-word",
              }}
            >
              {content.value}
            </pre>
          </div>
        )}
      </div>
    );
  },
});

export default new TextPreviewAdapter();
