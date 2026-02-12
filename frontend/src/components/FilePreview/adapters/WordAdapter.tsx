import { defineComponent, ref, h } from "vue";
import { Spin } from "ant-design-vue";
import VueOfficeDocx from "@vue-office/docx";
import "@vue-office/docx/lib/index.css";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// Word 预览适配器
class WordPreviewAdapter implements IPreviewAdapter {
  name = "WordPreviewAdapter";

  private supportedFormats = ["docx"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <WordPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// Word 预览组件
const WordPreview = defineComponent({
  name: "WordPreview",
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
      console.log("Word 渲染完成");
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = (err: any) => {
      console.error("Word 加载失败:", err);
      loading.value = false;
      error.value = true;
      props.onError?.(err);
    };

    return () => (
      <div class="word-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载 Word 文档中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">📝</div>
            <div class="error-text">Word 文档加载失败</div>
            <div class="error-hint">请检查文件格式或网络连接</div>
          </div>
        ) : (
          <div
            style={{
              display: loading.value ? "none" : "block",
              height: "80vh",
            }}
          >
            {h(VueOfficeDocx, {
              src: props.file.file_url,
              onRendered: handleRendered,
              onError: handleError,
            })}
          </div>
        )}
      </div>
    );
  },
});

export default new WordPreviewAdapter();
