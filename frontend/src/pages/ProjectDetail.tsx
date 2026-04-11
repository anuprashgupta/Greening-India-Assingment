import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Plus,
  Pencil,
  Trash2,
  AlertCircle,
  RefreshCw,
  ClipboardList,
  Save,
  X,
} from 'lucide-react';
import { getProject, updateProject, deleteProject } from '../api/projects';
import { createTask, updateTask, deleteTask } from '../api/tasks';
import { useAuth } from '../contexts/AuthContext';
import Navbar from '../components/Navbar';
import TaskCard from '../components/TaskCard';
import TaskModal from '../components/TaskModal';
import type { TaskFormData } from '../components/TaskModal';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import type { Project, Task } from '../types';

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [project, setProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [assigneeFilter, setAssigneeFilter] = useState<string>('all');

  // Task modal
  const [taskModalOpen, setTaskModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  // Project editing
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const isOwner = project && user && project.owner_id === user.id;

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message });
  };

  useEffect(() => {
    if (toast) {
      const timer = setTimeout(() => setToast(null), 3000);
      return () => clearTimeout(timer);
    }
  }, [toast]);

  const fetchData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError('');
    try {
      const proj = await getProject(id);
      setProject(proj);
      setTasks(proj.tasks || []);
    } catch {
      setError('Failed to load project. Please try again.');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Filtered tasks
  const filteredTasks = tasks.filter((t) => {
    if (statusFilter !== 'all' && t.status !== statusFilter) return false;
    if (assigneeFilter !== 'all' && (t.assignee_id || 'unassigned') !== assigneeFilter) return false;
    return true;
  });

  // Unique assignees for filter
  const assignees = Array.from(
    new Map(
      tasks
        .filter((t) => t.assignee_id)
        .map((t) => [t.assignee_id, t.assignee_name || t.assignee_id])
    ).entries()
  );

  // Project CRUD
  const handleEditProject = () => {
    if (!project) return;
    setEditName(project.name);
    setEditDesc(project.description || '');
    setEditing(true);
  };

  const handleSaveProject = async () => {
    if (!id || !editName.trim()) return;
    setSaving(true);
    try {
      const updated = await updateProject(id, { name: editName.trim(), description: editDesc.trim() });
      setProject(updated);
      setEditing(false);
      showToast('success', 'Project updated');
    } catch {
      showToast('error', 'Failed to update project');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteProject = async () => {
    if (!id) return;
    if (!window.confirm('Are you sure you want to delete this project? All tasks will be lost.')) return;
    setDeleting(true);
    try {
      await deleteProject(id);
      navigate('/projects', { replace: true });
    } catch {
      showToast('error', 'Failed to delete project');
      setDeleting(false);
    }
  };

  // Task CRUD
  const handleCreateTask = async (data: TaskFormData) => {
    if (!id) return;
    const newTask = await createTask(id, data);
    setTasks((prev) => [newTask, ...prev]);
    showToast('success', 'Task created');
  };

  const handleUpdateTask = async (data: TaskFormData) => {
    if (!editingTask) return;
    const updated = await updateTask(editingTask.id, data);
    setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)));
    showToast('success', 'Task updated');
  };

  const handleDeleteTask = async (task: Task) => {
    if (!window.confirm(`Delete task "${task.title}"?`)) return;
    try {
      await deleteTask(task.id);
      setTasks((prev) => prev.filter((t) => t.id !== task.id));
      showToast('success', 'Task deleted');
    } catch {
      showToast('error', 'Failed to delete task');
    }
  };

  // Optimistic status change
  const handleStatusChange = async (task: Task, newStatus: Task['status']) => {
    const oldStatus = task.status;
    // Optimistic update
    setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, status: newStatus } : t)));
    try {
      await updateTask(task.id, { status: newStatus });
    } catch {
      // Revert on error
      setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, status: oldStatus } : t)));
      showToast('error', 'Failed to update status');
    }
  };

  const openEditTask = (task: Task) => {
    setEditingTask(task);
    setTaskModalOpen(true);
  };

  const closeTaskModal = () => {
    setTaskModalOpen(false);
    setEditingTask(null);
  };

  const selectClass =
    'rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300';

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <Navbar />
        <div className="py-20">
          <LoadingSpinner size={40} text="Loading project..." />
        </div>
      </div>
    );
  }

  if (error || !project) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <Navbar />
        <div className="flex flex-col items-center justify-center py-20">
          <AlertCircle size={48} className="mb-3 text-red-400" />
          <p className="text-gray-600 dark:text-gray-400 mb-4">{error || 'Project not found'}</p>
          <div className="flex gap-3">
            <button
              onClick={() => navigate('/projects')}
              className="flex items-center gap-2 rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
            >
              <ArrowLeft size={16} />
              Back to Projects
            </button>
            <button
              onClick={fetchData}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <RefreshCw size={16} />
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <Navbar />

      <main className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Back button */}
        <button
          onClick={() => navigate('/projects')}
          className="mb-6 flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
        >
          <ArrowLeft size={16} />
          Back to Projects
        </button>

        {/* Project header */}
        <div className="mb-8 rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-800">
          {editing ? (
            <div className="space-y-3">
              <input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-lg font-bold text-gray-900 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              />
              <textarea
                value={editDesc}
                onChange={(e) => setEditDesc(e.target.value)}
                rows={2}
                className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 resize-none"
              />
              <div className="flex gap-2">
                <button
                  onClick={handleSaveProject}
                  disabled={saving}
                  className="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  <Save size={14} />
                  {saving ? 'Saving...' : 'Save'}
                </button>
                <button
                  onClick={() => setEditing(false)}
                  className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
                >
                  <X size={14} />
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div className="flex items-start justify-between gap-4">
              <div>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{project.name}</h1>
                {project.description && (
                  <p className="mt-1 text-gray-500 dark:text-gray-400">{project.description}</p>
                )}
              </div>
              {isOwner && (
                <div className="flex shrink-0 gap-2">
                  <button
                    onClick={handleEditProject}
                    className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
                  >
                    <Pencil size={14} />
                    Edit
                  </button>
                  <button
                    onClick={handleDeleteProject}
                    disabled={deleting}
                    className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20 disabled:opacity-50"
                  >
                    <Trash2 size={14} />
                    {deleting ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Tasks section */}
        <div className="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Tasks ({filteredTasks.length})
          </h2>
          <div className="flex flex-wrap items-center gap-3">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className={selectClass}
            >
              <option value="all">All Status</option>
              <option value="todo">To Do</option>
              <option value="in_progress">In Progress</option>
              <option value="done">Done</option>
            </select>

            <select
              value={assigneeFilter}
              onChange={(e) => setAssigneeFilter(e.target.value)}
              className={selectClass}
            >
              <option value="all">All Assignees</option>
              <option value="unassigned">Unassigned</option>
              {assignees.map(([id, name]) => (
                <option key={id} value={id!}>
                  {name}
                </option>
              ))}
            </select>

            <button
              onClick={() => {
                setEditingTask(null);
                setTaskModalOpen(true);
              }}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 transition-colors"
            >
              <Plus size={16} />
              Add Task
            </button>
          </div>
        </div>

        {filteredTasks.length === 0 ? (
          <EmptyState
            icon={<ClipboardList size={64} />}
            title={tasks.length === 0 ? 'No tasks yet' : 'No matching tasks'}
            description={
              tasks.length === 0
                ? 'Create your first task to get started.'
                : 'Try adjusting your filters to see more tasks.'
            }
            action={
              tasks.length === 0 ? (
                <button
                  onClick={() => {
                    setEditingTask(null);
                    setTaskModalOpen(true);
                  }}
                  className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
                >
                  <Plus size={16} />
                  Create Task
                </button>
              ) : undefined
            }
          />
        ) : (
          <div className="space-y-3">
            {filteredTasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onEdit={openEditTask}
                onDelete={handleDeleteTask}
                onStatusChange={handleStatusChange}
              />
            ))}
          </div>
        )}
      </main>

      <TaskModal
        open={taskModalOpen}
        task={editingTask}
        onClose={closeTaskModal}
        onSubmit={editingTask ? handleUpdateTask : handleCreateTask}
      />

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
