import { defineComponent, ref, h } from "vue";
import { Spin } from "ant-design-vue";
import VueOfficeExcel from "@vue-office/excel";
import "@vue-office/excel/lib/index.css";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// Excel 预览适配器
class ExcelPreviewAdapter implements IPreviewAdapter {
  name = "ExcelPreviewAdapter";

  private supportedFormats = ["xlsx", "xls"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <ExcelPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// Excel 预览组件
const ExcelPreview = defineComponent({
  name: "ExcelPreview",
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
      console.log("Excel 渲染完成");
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = (err: any) => {
      console.error("Excel 加载失败:", err);
      loading.value = false;
      error.value = true;
      props.onError?.(err);
    };

    return () => {
      // Excel 配置选项（在 render 函数中动态计算）
      const options = {
        xls: props.file.format?.toLowerCase() === "xls", // xls 文件设为 true，xlsx 设为 false
        minColLength: 0,
        minRowLength: 0,
        widthOffset: 10,
        heightOffset: 10,
      };

      return (
        <div class="excel-preview-container">
          {loading.value && (
            <div class="preview-loading">
              <Spin size="large" tip="加载 Excel 中..." />
            </div>
          )}

          {error.value ? (
            <div class="preview-error">
              <div class="error-icon">📊</div>
              <div class="error-text">Excel 加载失败</div>
              <div class="error-hint">请检查文件格式或网络连接</div>
            </div>
          ) : (
            <div
              style={{
                display: loading.value ? "none" : "block",
                height: "80vh",
              }}
            >
              {h(VueOfficeExcel, {
                src: props.file.file_url,
                options: options,
                onRendered: handleRendered,
                onError: handleError,
              })}
            </div>
          )}
        </div>
      );
    };
  },
});

export default new ExcelPreviewAdapter();
