import { Tag } from 'antd';

const STATUS_META: Record<string, { color: string; text: string }> = {
  on: { color: 'success', text: '开启' },
  off: { color: 'default', text: '关闭' },
  online: { color: 'success', text: '在线' },
  handled: { color: 'success', text: '已确认' },
  recovered: { color: 'cyan', text: '已恢复' },
  pending: { color: 'warning', text: '待处理' },
  warning: { color: 'warning', text: '预警' },
  critical: { color: 'error', text: '严重' },
};

export default function StatusTag({ status }: { status: string }) {
  const meta = STATUS_META[status] ?? { color: 'default', text: status };
  return <Tag color={meta.color}>{meta.text}</Tag>;
}
