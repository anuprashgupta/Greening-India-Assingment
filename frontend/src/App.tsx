import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Login from './pages/Login';
import Register from './pages/Register';
import Projects from './pages/Projects';
import ProjectDetail from './pages/ProjectDetail';

export default function App() {
  const { token } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={token ? <Navigate to="/projects" replace /> : <Login />} />
      <Route path="/register" element={token ? <Navigate to="/projects" replace /> : <Register />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/projects" element={<Projects />} />
        <Route path="/projects/:id" element={<ProjectDetail />} />
      </Route>
      <Route path="*" element={<Navigate to={token ? '/projects' : '/login'} replace />} />
    </Routes>
  );
}
