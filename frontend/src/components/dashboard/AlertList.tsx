import { Button, Card, List, Space, Tag, Typography, message } from 'antd';
import type { Alert } from '../../types/domain';
import StatusTag from '../common/StatusTag';
import { handleAlert } from '../../api/monitoring';

const { Text } = Typography;

export default function AlertList({ alerts, onUpdated }: { alerts: Alert[]; onUpdated: () => void }) {
  const acknowledge = async (id: number) => {
    try {
      await handleAlert(id);
      message.success('报警已确认');
      onUpdated();
    } catch (err: any) {
      // 重复确认或确认已恢复报警属于业务冲突：刷新列表保持与后端一致，失败不改本地状态。
      message.error(err?.response?.data?.message ?? '确认失败，请稍后重试');
      onUpdated();
    }
  };
  return (
    <Card title="实时报警" extra={<Button size="small" onClick={onUpdated}>刷新</Button>}>
      <List
        size="small"
        locale={{ emptyText: '当前没有报警' }}
        dataSource={alerts.slice(0, 6)}
        renderItem={(item) => (
          <List.Item actions={item.status === 'pending' ? [<Button type="link" key="ack" onClick={() => acknowledge(item.id)}>确认</Button>] : []}>
            <List.Item.Meta
              title={<Space size={4} wrap><StatusTag status={item.level} /><StatusTag status={item.status} /><span>{item.message}</span></Space>}
              description={
                <Space size={8} wrap>
                  <Text type="secondary">最新值 {item.value}</Text>
                  <Tag color="orange">连续超限 {item.count} 次</Tag>
                  <Text type="secondary">开始 {new Date(item.createdAt).toLocaleString('zh-CN')}</Text>
                  {item.handledAt && <Text type="success">已确认 {new Date(item.handledAt).toLocaleString('zh-CN')}</Text>}
                  {item.status === 'recovered' && item.recoveredAt && (
                    <Text type="success">已恢复（恢复值 {item.recoveredValue ?? item.value}）· {new Date(item.recoveredAt).toLocaleString('zh-CN')}</Text>
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
