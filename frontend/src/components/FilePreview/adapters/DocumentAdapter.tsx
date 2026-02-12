import { defineComponent, ref, onMounted } from "vue";
import { Spin, Button } from "ant-design-vue";
import { DownloadOutlined } from "@ant-design/icons-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// 文档预览适配器
class DocumentPreviewAdapter implements IPreviewAdapter {
  name = "DocumentPreviewAdapter";

  private supportedFormats = ["pdf", "txt", "md"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <DocumentPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 文档预览组件
const DocumentPreview = defineComponent({
  name: "DocumentPreview",
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
    const content = ref("");
    const isPdf = ref(false);

    onMounted(async () => {
      const format = props.file.format.toLowerCase();
      isPdf.value = format === "pdf";

      if (isPdf.value) {
        // PDF 使用 iframe 预览
        loading.value = false;
        props.onLoad?.();
      } else {
        // 文本文件，获取内容
        try {
          const response = await fetch(props.file.file_url);
          if (!response.ok) throw new Error("文件加载失败");

          content.value = await response.text();
          loading.value = false;
          props.onLoad?.();
        } catch (err) {
          loading.value = false;
          error.value = true;
          props.onError?.(err as Error);
        }
      }
    });

    const handleDownload = () => {
      window.open(props.file.file_url, "_blank");
    };

    return () => (
      <div class="document-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">📄</div>
            <div class="error-text">文档加载失败</div>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              onClick={handleDownload}
              style={{ marginTop: "16px" }}
            >
              下载文件
            </Button>
          </div>
        ) : isPdf.value ? (
          <iframe
            src={props.file.file_url}
            style={{
              width: "100%",
              height: "80vh",
              border: "none",
              display: loading.value ? "none" : "block",
            }}
            title={props.file.name}
          />
        ) : (
          <div class="text-preview">
            <div class="text-preview-header">
              <span class="file-name">{props.file.name}</span>
              <Button
                size="small"
                icon={<DownloadOutlined />}
                onClick={handleDownload}
              >
                下载
              </Button>
            </div>
            <pre class="text-content">{content.value}</pre>
          </div>
        )}
      </div>
    );
  },
});

export default new DocumentPreviewAdapter();
