import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import { Empty } from 'antd';
import type { EnvironmentReport } from '../../types/domain';

export default function ReportChart({ report }: { report: EnvironmentReport }) {
  const ref = useRef<HTMLDivElement>(null);
  const values = Object.entries(report.metrics || {});

  useEffect(() => {
    if (!ref.current || values.length === 0) return;
    const chart = echarts.init(ref.current);
    const maxValue = Math.max(...values.map(([, metric]) => metric.max)) * 1.2 || 1;
    chart.setOption({
      tooltip: {},
      radar: {
        indicator: values.map(([key]) => ({ name: key, max: maxValue })),
      },
      series: [
        {
          type: 'radar',
          data: [{ value: values.map(([, metric]) => metric.average), name: '平均值' }],
        },
      ],
    });
    const resize = () => chart.resize();
    window.addEventListener('resize', resize);
    return () => {
      window.removeEventListener('resize', resize);
      chart.dispose();
    };
  }, [report]);

  if (values.length === 0) {
    return <Empty description="当前时间范围内暂无数据" style={{ padding: 48 }} />;
  }

  return <div ref={ref} style={{ height: 260 }} aria-label="环境分析雷达图" />;
}
