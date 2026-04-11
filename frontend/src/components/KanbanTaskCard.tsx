import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Calendar, GripVertical, Pencil, Trash2, User } from 'lucide-react';
import PriorityBadge from './PriorityBadge';
import type { Task } from '../types';

interface Props {
  task: Task;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  isDragOverlay?: boolean;
}

export default function KanbanTaskCard({ task, onEdit, onDelete, isDragOverlay }: Props) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: task.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const formattedDate = task.due_date
    ? new Date(task.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    : null;

  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done';

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`group rounded-lg border bg-white dark:bg-gray-800 p-3 shadow-sm transition-all ${
        isDragging && !isDragOverlay
          ? 'opacity-30 border-dashed border-blue-300 dark:border-blue-600'
          : 'border-gray-200 dark:border-gray-700 hover:shadow-md'
      } ${isDragOverlay ? 'shadow-xl border-blue-400 dark:border-blue-500' : ''}`}
    >
      <div className="flex items-start gap-2">
        <button
          {...attributes}
          {...listeners}
          className="mt-0.5 shrink-0 cursor-grab rounded p-0.5 text-gray-300 hover:text-gray-500 dark:text-gray-600 dark:hover:text-gray-400 active:cursor-grabbing"
          aria-label="Drag to reorder"
        >
          <GripVertical size={14} />
        </button>

        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-gray-900 dark:text-white leading-snug">
            {task.title}
          </p>

          {task.description && (
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              {task.description}
            </p>
          )}

          <div className="mt-2 flex flex-wrap items-center gap-1.5">
            <PriorityBadge priority={task.priority} />

            {task.assignee_name && (
              <span className="inline-flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
                <User size={10} />
                {task.assignee_name}
              </span>
            )}

            {formattedDate && (
              <span
                className={`inline-flex items-center gap-1 text-xs ${
                  isOverdue ? 'text-red-500' : 'text-gray-400 dark:text-gray-500'
                }`}
              >
                <Calendar size={10} />
                {formattedDate}
              </span>
            )}
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => onEdit(task)}
            className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
            aria-label="Edit task"
          >
            <Pencil size={12} />
          </button>
          <button
            onClick={() => onDelete(task)}
            className="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20 dark:hover:text-red-400"
            aria-label="Delete task"
          >
            <Trash2 size={12} />
          </button>
        </div>
      </div>
    </div>
  );
}
