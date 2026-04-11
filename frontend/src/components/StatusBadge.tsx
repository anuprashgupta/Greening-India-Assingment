interface Props {
  status: 'todo' | 'in_progress' | 'done';
}

const config = {
  todo: { label: 'To Do', classes: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300' },
  in_progress: { label: 'In Progress', classes: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300' },
  done: { label: 'Done', classes: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300' },
};

export default function StatusBadge({ status }: Props) {
  const { label, classes } = config[status];
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${classes}`}>
      {label}
    </span>
  );
}
