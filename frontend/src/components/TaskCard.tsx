import { Calendar, User, Pencil, Trash2 } from 'lucide-react';
import PriorityBadge from './PriorityBadge';
import type { Task } from '../types';

interface Props {
  task: Task;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onStatusChange: (task: Task, status: Task['status']) => void;
}

export default function TaskCard({ task, onEdit, onDelete, onStatusChange }: Props) {
  const formattedDate = task.due_date
    ? new Date(task.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : null;

  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done';

  return (
    <div className="group rounded-lg border border-gray-200 bg-white p-4 transition-all hover:shadow-sm dark:border-gray-700 dark:bg-gray-800">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <h4 className="font-medium text-gray-900 dark:text-white">{task.title}</h4>
          {task.description && (
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">{task.description}</p>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => onEdit(task)}
            className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
            aria-label="Edit task"
          >
            <Pencil size={14} />
          </button>
          <button
            onClick={() => onDelete(task)}
            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20 dark:hover:text-red-400"
            aria-label="Delete task"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <select
          value={task.status}
          onChange={(e) => onStatusChange(task, e.target.value as Task['status'])}
          className={`rounded-full border-0 px-2.5 py-0.5 text-xs font-medium cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 ${
            task.status === 'todo'
              ? 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
              : task.status === 'in_progress'
              ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
              : 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300'
          }`}
          aria-label="Change status"
        >
          <option value="todo">To Do</option>
          <option value="in_progress">In Progress</option>
          <option value="done">Done</option>
        </select>
        <PriorityBadge priority={task.priority} />

        {task.assignee_name && (
          <span className="inline-flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
            <User size={12} />
            {task.assignee_name}
          </span>
        )}

        {formattedDate && (
          <span
            className={`inline-flex items-center gap-1 text-xs ${
              isOverdue ? 'text-red-500' : 'text-gray-500 dark:text-gray-400'
            }`}
          >
            <Calendar size={12} />
            {formattedDate}
          </span>
        )}
      </div>
    </div>
  );
}
