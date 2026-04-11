import client from './client';
import type { Project, Task } from '../types';

interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total_count: number;
  total_pages: number;
}

interface ProjectWithTasks extends Project {
  tasks: Task[];
}

export async function getProjects(): Promise<Project[]> {
  const { data } = await client.get<PaginatedResponse<Project>>('/projects');
  return data.data;
}

export async function getProject(id: string): Promise<ProjectWithTasks> {
  const { data } = await client.get<ProjectWithTasks>(`/projects/${id}`);
  return data;
}

export async function createProject(payload: { name: string; description: string }): Promise<Project> {
  const { data } = await client.post<Project>('/projects', payload);
  return data;
}

export async function updateProject(id: string, payload: { name?: string; description?: string }): Promise<Project> {
  const { data } = await client.patch<Project>(`/projects/${id}`, payload);
  return data;
}

export async function deleteProject(id: string): Promise<void> {
  await client.delete(`/projects/${id}`);
}
