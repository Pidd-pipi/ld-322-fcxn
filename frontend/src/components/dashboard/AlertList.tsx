import { Button, Card, List, Space, Tag, Typography, message } from 'antd';
import axios from 'axios';
import type { Alert } from '../../types/domain';
import StatusTag from '../common/StatusTag';
import { handleAlert } from '../../api/monitoring';

const { Text } = Typography;

const formatTime = (value?: string) => (value ? new Date(value).toLocaleString('zh-CN') : '—');

export default function AlertList({ alerts, onUpdated }: { alerts: Alert[]; onUpdated: () => void }) {
  const resolve = async (id: number) => {
    try {
      await handleAlert(id);
      message.success('报警已确认');
      onUpdated();
    } catch (err) {
      const detail = axios.isAxiosError(err) ? err.response?.data?.message : undefined;
      message.error(detail ?? '确认失败，请先登录演示账号');
    }
  };

  return (
    <Card
      title="实时报警"
      extra={<Button size="small" onClick={onUpdated}>刷新</Button>}
    >
      <List
        size="small"
        locale={{ emptyText: '当前没有报警' }}
        dataSource={alerts.slice(0, 6)}
        renderItem={(item) => (
          <List.Item
            actions={item.status === 'pending'
              ? [<Button type="link" key="handle" onClick={() => resolve(item.id)}>确认报警</Button>]
              : []}
          >
            <List.Item.Meta
              title={
                <Space size={4} wrap>
                  <StatusTag status={item.level} />
                  <StatusTag status={item.status} />
                  <Tag>连续 {item.occurrenceCount ?? 1} 次超限</Tag>
                  <span>{item.message}</span>
                </Space>
              }
              description={
                <Space direction="vertical" size={0}>
                  <Text type="secondary">触发：{formatTime(item.createdAt)} · 最新值 {item.value}{item.sensor?.unit ?? ''}</Text>
                  {item.status === 'recovered' && (
                    <Text type="success">
                      已自动恢复 · 恢复值 {item.recoveredValue ?? '—'}{item.sensor?.unit ?? ''} · 恢复时间 {formatTime(item.recoveredAt)}
                    </Text>
                  )}
                  {item.status === 'handled' && (
                    <Text type="secondary">人工确认时间：{formatTime(item.acknowledgedAt ?? item.handledAt)}</Text>
                  )}
                </Space>
              }
            />
          </List.Item>
        )}
      />
    </Card>
  );
}
