import axios from 'axios';

// Create a new axios instance with the base URL of your backend
const API = axios.create({
  baseURL: 'http://localhost:8080',
});

// This is an "interceptor". It's a special function that runs
// automatically before every single API request is sent.
API.interceptors.request.use((config) => {
  // We only run this code in the browser, not on the server
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('token');
    if (token) {
      // If a token is found in localStorage, add it to the
      // 'Authorization' header of the request.
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});


// --- AUTH ENDPOINTS ---
export const login = (email: string, password: string) =>
  API.post('/auth/login', { email, password });

export const register = (email: string, password: string) =>
  API.post('/auth/register', { email, password });


// --- TASK ENDPOINTS ---
// Define a type for our Task for better type safety
interface TaskPayload {
    title: string;
    description: string;
}

interface TaskUpdatePayload {
    status?: 'Pending' | 'Done';
}

export const getTasks = () => API.get('/tasks');
export const createTask = (task: TaskPayload) => API.post('/tasks', task);
export const updateTask = (id: number, task: TaskUpdatePayload) => API.put(`/tasks/${id}`, task);
export const deleteTask = (id: number) => API.delete(`/tasks/${id}`);