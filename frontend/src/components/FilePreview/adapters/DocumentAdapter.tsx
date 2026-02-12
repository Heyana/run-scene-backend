import { defineComponent, ref, h } from "vue";
import { Spin } from "ant-design-vue";
import VueOfficePdf from "@vue-office/pdf";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// PDF 文档预览适配器
class DocumentPreviewAdapter implements IPreviewAdapter {
  name = "DocumentPreviewAdapter";

  private supportedFormats = ["pdf"];

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

// PDF 预览组件
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

    const handleRendered = () => {
      console.log("PDF 渲染完成");
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = (err: any) => {
      console.error("PDF 加载失败:", err);
      loading.value = false;
      error.value = true;
      props.onError?.(err);
    };

    return () => (
      <div
        class="document-preview-container"
        style={{ height: "80vh", overflow: "hidden" }}
      >
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载 PDF 中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">📄</div>
            <div class="error-text">PDF 加载失败</div>
            <div class="error-hint">请检查文件格式或网络连接</div>
          </div>
        ) : (
          h(VueOfficePdf, {
            src: props.file.file_url,
            onRendered: handleRendered,
            onError: handleError,
          })
        )}
      </div>
    );
  },
});

export default new DocumentPreviewAdapter();
