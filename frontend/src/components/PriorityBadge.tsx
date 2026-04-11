interface Props {
  priority: 'low' | 'medium' | 'high';
}

const config = {
  low: { label: 'Low', classes: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300' },
  medium: { label: 'Medium', classes: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300' },
  high: { label: 'High', classes: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300' },
};

export default function PriorityBadge({ priority }: Props) {
  const { label, classes } = config[priority];
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${classes}`}>
      {label}
    </span>
  );
}
