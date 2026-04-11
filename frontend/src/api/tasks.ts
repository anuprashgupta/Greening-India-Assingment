import client from './client';
import type { Task } from '../types';

interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total_count: number;
  total_pages: number;
}

export interface CreateTaskPayload {
  title: string;
  description?: string;
  status?: 'todo' | 'in_progress' | 'done';
  priority?: 'low' | 'medium' | 'high';
  assignee_id?: string | null;
  due_date?: string | null;
}

export interface UpdateTaskPayload extends Partial<CreateTaskPayload> {}

export async function getTasks(projectId: string, params?: { status?: string; assignee?: string }): Promise<Task[]> {
  const { data } = await client.get<PaginatedResponse<Task>>(`/projects/${projectId}/tasks`, { params });
  return data.data;
}

export async function createTask(projectId: string, payload: CreateTaskPayload): Promise<Task> {
  const { data } = await client.post<Task>(`/projects/${projectId}/tasks`, payload);
  return data;
}

export async function updateTask(taskId: string, payload: UpdateTaskPayload): Promise<Task> {
  const { data } = await client.patch<Task>(`/tasks/${taskId}`, payload);
  return data;
}

export async function deleteTask(taskId: string): Promise<void> {
  await client.delete(`/tasks/${taskId}`);
}
