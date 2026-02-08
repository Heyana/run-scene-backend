import { defineComponent, type PropType } from "vue";
import { Button, Input, Select } from "ant-design-vue";
import { SearchOutlined, HomeOutlined } from "@ant-design/icons-vue";
import { useRouter } from "vue-router";
import "./ResourceHeader.less";

const { Search } = Input;

export interface StatItem {
  icon: any;
  label: string;
  value: string | number;
  color?: string;
}

export interface ActionButton {
  label: string;
  icon?: any;
  type?: "default" | "primary" | "dashed" | "link" | "text";
  loading?: boolean;
  onClick: () => void;
}

export interface FilterOption {
  label: string;
  value: any;
}

export interface ResourceHeaderProps {
  stats?: StatItem[];
  actions?: ActionButton[];
  showHomeButton?: boolean;
  onSearch?: (value: string) => void;
  searchPlaceholder?: string;
  filters?: {
    label: string;
    value: any;
    options: FilterOption[];
    onChange: (value: any) => void;
  }[];
  pageSizeOptions?: number[];
  pageSize?: number;
  onPageSizeChange?: (value: number) => void;
}

export default defineComponent({
  name: "ResourceHeader",
  props: {
    stats: {
      type: Array as PropType<StatItem[]>,
      default: () => [],
    },
    showHomeButton: {
      type: Boolean,
      default: true,
    },
    actions: {
      type: Array as PropType<ActionButton[]>,
      default: () => [],
    },
    onSearch: {
      type: Function as PropType<(value: string) => void>,
    },
    searchPlaceholder: {
      type: String,
      default: "搜索",
    },
    filters: {
      type: Array as PropType<ResourceHeaderProps["filters"]>,
      default: () => [],
    },
    pageSizeOptions: {
      type: Array as PropType<number[]>,
      default: () => [12, 24, 48, 96],
    },
    pageSize: {
      type: Number,
      default: 12,
    },
    onPageSizeChange: {
      type: Function as PropType<(value: number) => void>,
    },
  },
  setup(props) {
    const router = useRouter();

    const goHome = () => {
      router.push("/");
    };

    return () => (
      <div class="resource-header">
        {/* 统计信息 */}
        {props.stats && props.stats.length > 0 && (
          <div class="stats-section">
            {props.stats.map((stat, index) => (
              <div key={index} class="stat-item">
                {stat.icon && (
                  <stat.icon
                    class="stat-icon"
                    style={{ color: stat.color || "#1890ff" }}
                  />
                )}
                <div class="stat-content">
                  <div class="stat-label">{stat.label}</div>
                  <div class="stat-value">{stat.value}</div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 操作区 */}
        <div class="actions-section">
          {/* 返回首页按钮 */}
          {props.showHomeButton && (
            <Button onClick={goHome}>
              {{
                icon: () => <HomeOutlined />,
                default: () => "返回首页",
              }}
            </Button>
          )}

          {/* 搜索 */}
          {props.onSearch && (
            <Search
              placeholder={props.searchPlaceholder}
              allowClear
              onSearch={props.onSearch}
              style={{ width: 240 }}
              v-slots={{
                enterButton: () => <SearchOutlined />,
              }}
            />
          )}

          {/* 过滤器 */}
          {props.filters?.map((filter, index) => (
            <Select
              key={index}
              placeholder={filter.label}
              allowClear
              value={filter.value}
              onChange={filter.onChange}
              style={{ width: 130 }}
              options={filter.options}
            />
          ))}

          {/* 分页大小 */}
          {props.onPageSizeChange && (
            <Select
              value={props.pageSize}
              onChange={(value) => props.onPageSizeChange?.(value as number)}
              style={{ width: 110 }}
              options={props.pageSizeOptions.map((size) => ({
                label: `${size} 条/页`,
                value: size,
              }))}
            />
          )}

          {/* 操作按钮 */}
          {props.actions?.map((action, index) => (
            <Button
              key={index}
              type={action.type || "default"}
              loading={action.loading}
              onClick={action.onClick}
            >
              {{
                icon: action.icon ? () => <action.icon /> : undefined,
                default: () => action.label,
              }}
            </Button>
          ))}
        </div>
      </div>
    );
  },
});
