import { useEffect, useState, useCallback } from 'react';
import { Plus, AlertCircle, RefreshCw } from 'lucide-react';
import { getProjects, createProject } from '../api/projects';
import ProjectCard from '../components/ProjectCard';
import CreateProjectModal from '../components/CreateProjectModal';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import Navbar from '../components/Navbar';
import type { Project } from '../types';

export default function Projects() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const fetchProjects = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getProjects();
      setProjects(data || []);
    } catch {
      setError('Failed to load projects. Please try again.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  useEffect(() => {
    if (toast) {
      const timer = setTimeout(() => setToast(null), 3000);
      return () => clearTimeout(timer);
    }
  }, [toast]);

  const handleCreate = async (data: { name: string; description: string }) => {
    const newProject = await createProject(data);
    setProjects((prev) => [newProject, ...prev]);
    setToast({ type: 'success', message: 'Project created successfully' });
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <Navbar />

      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Projects</h1>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Manage your projects and tasks
            </p>
          </div>
          <button
            onClick={() => setModalOpen(true)}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm hover:bg-blue-700 transition-colors"
          >
            <Plus size={18} />
            New Project
          </button>
        </div>

        {loading ? (
          <div className="py-20">
            <LoadingSpinner size={40} text="Loading projects..." />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-20">
            <AlertCircle size={48} className="mb-3 text-red-400" />
            <p className="text-gray-600 dark:text-gray-400 mb-4">{error}</p>
            <button
              onClick={fetchProjects}
              className="flex items-center gap-2 rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
            >
              <RefreshCw size={16} />
              Retry
            </button>
          </div>
        ) : projects.length === 0 ? (
          <EmptyState
            title="No projects yet"
            description="Create your first project to start organizing tasks."
            action={
              <button
                onClick={() => setModalOpen(true)}
                className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-blue-700"
              >
                <Plus size={18} />
                Create Project
              </button>
            }
          />
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((p) => (
              <ProjectCard key={p.id} project={p} />
            ))}
          </div>
        )}
      </main>

      <CreateProjectModal open={modalOpen} onClose={() => setModalOpen(false)} onSubmit={handleCreate} />

      {/* Toast notification */}
      {toast && (
        <div
          className={`fixed bottom-6 right-6 z-50 rounded-lg px-4 py-3 text-sm font-medium text-white shadow-lg transition-all ${
            toast.type === 'success' ? 'bg-green-600' : 'bg-red-600'
          }`}
        >
          {toast.message}
        </div>
      )}
    </div>
  );
}
