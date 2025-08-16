import { useState, useEffect, FormEvent, CSSProperties } from 'react';
import { useRouter } from 'next/router';
import { getTasks, createTask, updateTask, deleteTask } from '../services/api';

// Define the structure of a Task object for TypeScript
interface Task {
  id: number;
  title: string;
  description: string;
  status: 'Pending' | 'Done';
}

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const router = useRouter();

  // This function fetches tasks from the backend.
  const fetchTasks = async () => {
    try {
      const response = await getTasks();
      setTasks(response.data);
    } catch (err: any) {
      setError('Failed to fetch tasks. Please try again.');
      if (err.response?.status === 401) {
        handleLogout();
      }
    } finally {
      setIsLoading(false);
    }
  };

  // useEffect runs when the component first loads.
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/');
    } else {
      fetchTasks();
    }
  }, [router]);

  // Handles the submission of the "Add Task" form
  const handleAddTask = async (e: FormEvent) => {
    e.preventDefault();
    if (!title) return;

    try {
      await createTask({ title, description });
      setTitle('');
      setDescription('');
      fetchTasks();
    } catch (err) {
      setError('Failed to create task.');
    }
  };

  // Handles changing a task's status
  const handleUpdateStatus = async (id: number, currentStatus: 'Pending' | 'Done') => {
    const newStatus = currentStatus === 'Pending' ? 'Done' : 'Pending';
    try {
      await updateTask(id, { status: newStatus });
      fetchTasks();
    } catch (err) {
      setError('Failed to update task status.');
    }
  };

  // Handles deleting a task
  const handleDeleteTask = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this task?')) {
      try {
        await deleteTask(id);
        fetchTasks();
      } catch (err) {
        setError('Failed to delete task.');
      }
    }
  };

  // Handles user logout
  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/');
  };

  if (isLoading) {
    return <div style={styles.loadingContainer}>Loading tasks...</div>;
  }

  return (
    <div style={styles.page}>
      <nav style={styles.nav}>
        <h1 style={styles.navTitle}>Task Tracker</h1>
        <button onClick={handleLogout} style={styles.logoutButton}>
          Logout
        </button>
      </nav>

      <main style={styles.main}>
        {/* Form to create a new task */}
        <div style={styles.card}>
          <h2 style={styles.formTitle}>Add a New Task</h2>
          <form onSubmit={handleAddTask}>
            <div style={styles.inputGroup}>
              <input
                type="text"
                placeholder="Task Title (e.g., Finish report)"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                style={styles.input}
              />
            </div>
            <div style={styles.inputGroup}>
              <textarea
                placeholder="Description (optional)"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                style={styles.textarea}
                rows={3}
              />
            </div>
            <button type="submit" style={styles.button}>
              Add Task
            </button>
          </form>
        </div>

        {error && <p style={styles.errorText}>{error}</p>}

        {/* List of existing tasks */}
        <div style={styles.taskList}>
          {tasks.length > 0 ? (
            tasks.map((task, index) => (
              <div key={task.id} style={{...styles.taskItem, borderTop: index > 0 ? '1px solid #e9ecef' : 'none'}}>
                <div>
                  <h3 style={{...styles.taskTitle, textDecoration: task.status === 'Done' ? 'line-through' : 'none', color: task.status === 'Done' ? '#6c757d' : '#212529'}}>
                    {task.title}
                  </h3>
                  {task.description && <p style={styles.taskDescription}>{task.description}</p>}
                </div>
                <div style={styles.taskActions}>
                  <button
                    onClick={() => handleUpdateStatus(task.id, task.status)}
                    style={{...styles.statusButton, backgroundColor: task.status === 'Pending' ? '#ffc107' : '#28a745'}}
                  >
                    {task.status}
                  </button>
                  <button
                    onClick={() => handleDeleteTask(task.id)}
                    style={styles.deleteButton}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div style={{...styles.card, textAlign: 'center'}}>
                <p style={styles.noTasksText}>You have no tasks yet. Add one above!</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

// Basic CSS-in-JS styles
const styles: { [key: string]: CSSProperties } = {
  page: {
    minHeight: '100vh',
    backgroundColor: '#f8f9fa',
    fontFamily: 'sans-serif',
    color: '#212529',
  },
  loadingContainer: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: '100vh',
  },
  nav: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '1rem',
    backgroundColor: 'rgba(255, 255, 255, 0.8)',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
    position: 'sticky',
    top: 0,
    zIndex: 10,
    borderBottom: '1px solid #e9ecef',
  },
  navTitle: {
    fontSize: '1.5rem',
    fontWeight: 'bold',
    color: '#4a5568',
  },
  logoutButton: {
    padding: '0.5rem 1rem',
    fontSize: '0.875rem',
    color: 'white',
    backgroundColor: '#dc3545',
    border: 'none',
    borderRadius: '0.375rem',
    cursor: 'pointer',
  },
  main: {
    padding: '1.5rem',
    margin: '0 auto',
    maxWidth: '56rem',
  },
  card: {
    padding: '1.5rem',
    marginBottom: '2rem',
    backgroundColor: '#ffffff',
    borderRadius: '0.5rem',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
  },
  formTitle: {
    marginBottom: '1rem',
    fontSize: '1.25rem',
    fontWeight: 600,
  },
  inputGroup: {
    marginBottom: '1rem',
  },
  input: {
    width: '100%',
    padding: '0.75rem',
    border: '1px solid #ced4da',
    borderRadius: '0.375rem',
  },
  textarea: {
    width: '100%',
    padding: '0.75rem',
    border: '1px solid #ced4da',
    borderRadius: '0.375rem',
    fontFamily: 'sans-serif',
  },
  button: {
    width: '100%',
    padding: '0.75rem',
    fontWeight: 600,
    color: 'white',
    backgroundColor: '#4a5568',
    border: 'none',
    borderRadius: '0.375rem',
    cursor: 'pointer',
  },
  errorText: {
    marginBottom: '1rem',
    textAlign: 'center',
    color: '#dc3545',
  },
  taskList: {
    backgroundColor: '#ffffff',
    borderRadius: '0.5rem',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
  },
  taskItem: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '1rem',
  },
  taskTitle: {
    fontSize: '1.125rem',
    fontWeight: 500,
  },
  taskDescription: {
    marginTop: '0.25rem',
    fontSize: '0.875rem',
    color: '#6c757d',
  },
  taskActions: {
    display: 'flex',
    alignItems: 'center',
    marginLeft: '1rem',
  },
  statusButton: {
    padding: '0.25rem 0.75rem',
    fontSize: '0.75rem',
    fontWeight: 600,
    color: 'white',
    borderRadius: '9999px',
    border: 'none',
    marginRight: '0.5rem',
    cursor: 'pointer',
  },
  deleteButton: {
    padding: '0.25rem 0.75rem',
    fontSize: '0.75rem',
    fontWeight: 600,
    color: 'white',
    backgroundColor: '#dc3545',
    borderRadius: '9999px',
    border: 'none',
    cursor: 'pointer',
  },
  noTasksText: {
    color: '#6c757d',
  },
};