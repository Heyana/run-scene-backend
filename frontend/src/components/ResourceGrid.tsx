import { defineComponent, type PropType } from "vue";
import { Spin, Empty, Pagination, Image } from "ant-design-vue";
import { PictureOutlined } from "@ant-design/icons-vue";
import "./ResourceGrid.less";
import type { JSX } from "vue/jsx-runtime";

export interface ResourceGridProps<T = any> {
  loading: boolean;
  data: T[];
  total: number;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number, pageSize: number) => void;
  renderPreview?: (item: T) => JSX.Element;
  renderContent?: (item: T) => JSX.Element;
  onCardClick?: (item: T) => void;
  onContextMenu?: (e: MouseEvent, item: T) => void;
}

export default defineComponent({
  name: "ResourceGrid",
  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    data: {
      type: Array as PropType<any[]>,
      required: true,
    },
    total: {
      type: Number,
      required: true,
    },
    currentPage: {
      type: Number,
      required: true,
    },
    pageSize: {
      type: Number,
      required: true,
    },
    onPageChange: {
      type: Function as PropType<(page: number, pageSize: number) => void>,
      required: true,
    },
    renderPreview: {
      type: Function as PropType<(item: any) => JSX.Element>,
    },
    renderContent: {
      type: Function as PropType<(item: any) => JSX.Element>,
    },
    onCardClick: {
      type: Function as PropType<(item: any) => void>,
    },
    onContextMenu: {
      type: Function as PropType<(e: MouseEvent, item: any) => void>,
    },
  },
  setup(props) {
    // 默认预览渲染
    const defaultRenderPreview = (item: any) => {
      const previewUrl = item.thumbnail_url || item.preview_url || "";
      if (previewUrl) {
        return (
          <Image
            src={previewUrl}
            width="100%"
            height="100%"
            style={{ objectFit: "cover" }}
            preview={{ src: previewUrl }}
          />
        );
      }
      return (
        <div class="preview-placeholder">
          <PictureOutlined />
        </div>
      );
    };

    // 默认内容渲染
    const defaultRenderContent = (item: any) => {
      return (
        <div class="resource-name" title={item.name}>
          {item.name || item.title || "未命名"}
        </div>
      );
    };

    return () => (
      <div class="resource-grid-container">
        <Spin spinning={props.loading}>
          {props.data.length === 0 ? (
            <div class="empty-container">
              <Empty description="暂无数据" />
            </div>
          ) : (
            <>
              <div class="resource-grid">
                {props.data.map((item: any) => (
                  <div
                    key={item.id}
                    class="resource-card"
                    onClick={() => props.onCardClick?.(item)}
                    onContextmenu={(e: MouseEvent) =>
                      props.onContextMenu?.(e, item)
                    }
                  >
                    {/* 预览图 */}
                    <div class="resource-preview">
                      {props.renderPreview
                        ? props.renderPreview(item)
                        : defaultRenderPreview(item)}
                    </div>

                    {/* 资源信息 */}
                    <div class="resource-info">
                      {props.renderContent
                        ? props.renderContent(item)
                        : defaultRenderContent(item)}
                    </div>
                  </div>
                ))}
              </div>

              {/* 分页 */}
              <div class="pagination-container">
                <Pagination
                  current={props.currentPage}
                  pageSize={props.pageSize}
                  total={props.total}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total: number) => `共 ${total} 条`}
                  onChange={props.onPageChange}
                />
              </div>
            </>
          )}
        </Spin>
      </div>
    );
  },
});
