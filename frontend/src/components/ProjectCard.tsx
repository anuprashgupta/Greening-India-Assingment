import { useNavigate } from 'react-router-dom';
import { FolderKanban, ListChecks } from 'lucide-react';
import type { Project } from '../types';

interface Props {
  project: Project;
}

export default function ProjectCard({ project }: Props) {
  const navigate = useNavigate();
  const taskCount = project.tasks?.length ?? 0;

  return (
    <button
      onClick={() => navigate(`/projects/${project.id}`)}
      className="group w-full text-left rounded-xl border border-gray-200 bg-white p-5 shadow-sm transition-all hover:shadow-md hover:border-blue-300 dark:border-gray-700 dark:bg-gray-800 dark:hover:border-blue-600"
    >
      <div className="flex items-start gap-3">
        <div className="mt-0.5 rounded-lg bg-blue-50 p-2 text-blue-500 dark:bg-blue-900/30">
          <FolderKanban size={20} />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 truncate">
            {project.name}
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
            {project.description || 'No description'}
          </p>
          <div className="mt-3 flex items-center gap-1.5 text-xs text-gray-400 dark:text-gray-500">
            <ListChecks size={14} />
            <span>{taskCount} task{taskCount !== 1 ? 's' : ''}</span>
          </div>
        </div>
      </div>
    </button>
  );
}
