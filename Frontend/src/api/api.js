import axios from 'axios';

// Base configuration for Magistrala API
const BASE_URL = 'http://localhost'; // For API calls (Local Magistrala via nginx)
const WEBSOCKET_URL = 'ws://localhost/ws'; // For WebSocket connection

// Create axios instance with default config
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Token management
let authToken = localStorage.getItem('magistrala_token');

// Demo mode data storage
const isDemoMode = () => authToken && authToken.startsWith('demo-token');

// Demo data storage in localStorage
const getDemoData = (key) => {
  const data = localStorage.getItem(`demo_${key}`);
  return data ? JSON.parse(data) : [];
};

const setDemoData = (key, data) => {
  localStorage.setItem(`demo_${key}`, JSON.stringify(data));
};

// Initialize demo data if not exists
const initDemoData = () => {
  if (!localStorage.getItem('demo_users')) {
    setDemoData('users', [
      {
        id: 'user-1',
        first_name: 'Admin',
        last_name: 'User',
        email: 'admin@example.com',
        status: 'enabled',
        role: 'admin',
        created_at: new Date().toISOString()
      },
      {
        id: 'user-2',
        first_name: 'Demo',
        last_name: 'User',
        email: 'demo@example.com',
        status: 'enabled',
        role: 'user',
        created_at: new Date().toISOString()
      }
    ]);
  }
  
  if (!localStorage.getItem('demo_things')) {
    setDemoData('things', [
      {
        id: 'device-1',
        name: 'Temperature Sensor',
        credentials: { secret: 'temp-sensor-secret' },
        status: 'enabled',
        created_at: new Date().toISOString(),
        metadata: { type: 'sensor', location: 'Room 1' }
      },
      {
        id: 'device-2', 
        name: 'Smart Light',
        credentials: { secret: 'light-secret' },
        status: 'enabled',
        created_at: new Date().toISOString(),
        metadata: { type: 'actuator', location: 'Room 2' }
      }
    ]);
  }
  
  if (!localStorage.getItem('demo_channels')) {
    setDemoData('channels', [
      {
        id: 'channel-1',
        name: 'Sensor Data',
        status: 'enabled',
        created_at: new Date().toISOString(),
        metadata: { type: 'data', protocol: 'mqtt' }
      },
      {
        id: 'channel-2',
        name: 'Control Commands',
        status: 'enabled', 
        created_at: new Date().toISOString(),
        metadata: { type: 'control', protocol: 'http' }
      }
    ]);
  }
};

const generateId = () => `demo-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    if (authToken) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('magistrala_token');
      authToken = null;
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API calls
export const auth = {
  login: async (email, password) => {
    try {
      // Use default credentials if none are provided
      if (!email || !password) {
        email = 'admin@example.com'; // Default email
        password = '12345678'; // Default password (matches Magistrala config)
      }
      // Demo mode - bypass authentication for now
      if (email === 'admin@example.com' && password === '12345678') {
        const demoToken = 'demo-token-' + Date.now();
        authToken = demoToken;
        localStorage.setItem('magistrala_token', demoToken);
        initDemoData(); // Initialize demo data
        return { success: true, token: demoToken };
      }
      
      const response = await api.post('/users/tokens/issue', { identity: email, secret: password });
      const token = response.data.access_token;
      authToken = token;
      localStorage.setItem('magistrala_token', token);
      return { success: true, token };
    } catch (error) {
      console.error('Login error:', error);
      let errorMessage = 'Login failed';
      if (error.code === 'ECONNREFUSED' || error.code === 'ERR_NETWORK') {
        errorMessage = 'Cannot connect to Magistrala server. Using demo mode with credentials: admin@example.com / 12345678';
      } else if (error.response) {
        errorMessage = error.response.data.message || 'An error occurred';
      } else {
        errorMessage = error.message || 'An unknown error occurred';
      }
      return { 
        success: false, 
        error: errorMessage,
        details: {
          status: error.response?.status,
          code: error.code,
          baseURL: BASE_URL
        }
      };
    }
  },

  logout: () => {
    authToken = null;
    localStorage.removeItem('magistrala_token');
  },

  isAuthenticated: () => {
    return !!authToken;
  },

  getToken: () => authToken
};

// Users API calls
export const users = {
  getAll: async (offset = 0, limit = 20) => {
    if (isDemoMode()) {
      const users = getDemoData('users');
      const paginatedUsers = users.slice(offset, offset + limit);
      return { 
        success: true, 
        data: { 
          users: paginatedUsers, 
          total: users.length,
          offset,
          limit 
        } 
      };
    }
    
    try {
      const response = await api.get(`/users?offset=${offset}&limit=${limit}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch users' 
      };
    }
  },

  create: async (userData) => {
    if (isDemoMode()) {
      const users = getDemoData('users');
      const newUser = {
        id: generateId(),
        ...userData,
        created_at: new Date().toISOString(),
        status: userData.status || 'enabled'
      };
      users.push(newUser);
      setDemoData('users', users);
      return { success: true, data: newUser };
    }
    
    try {
      const response = await api.post('/users', userData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to create user' 
      };
    }
  },

  getById: async (id) => {
    try {
      const response = await api.get(`/users/${id}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch user' 
      };
    }
  },

  update: async (id, userData) => {
    try {
      const response = await api.put(`/users/${id}`, userData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to update user' 
      };
    }
  },

  delete: async (id) => {
    if (isDemoMode()) {
      const users = getDemoData('users');
      const filteredUsers = users.filter(user => user.id !== id);
      setDemoData('users', filteredUsers);
      return { success: true };
    }
    
    try {
      await api.delete(`/users/${id}`);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to delete user' 
      };
    }
  }
};

// Things (Devices) API calls
export const things = {
  getAll: async (offset = 0, limit = 20) => {
    if (isDemoMode()) {
      const things = getDemoData('things');
      const paginatedThings = things.slice(offset, offset + limit);
      return { 
        success: true, 
        data: { 
          things: paginatedThings, 
          total: things.length,
          offset,
          limit 
        } 
      };
    }
    
    try {
      const response = await api.get(`/things?offset=${offset}&limit=${limit}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch devices' 
      };
    }
  },

  create: async (thingData) => {
    if (isDemoMode()) {
      const things = getDemoData('things');
      const newThing = {
        id: generateId(),
        ...thingData,
        credentials: { secret: generateId() },
        created_at: new Date().toISOString(),
        status: thingData.status || 'enabled'
      };
      things.push(newThing);
      setDemoData('things', things);
      return { success: true, data: newThing };
    }
    
    try {
      const response = await api.post('/things', thingData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to create device' 
      };
    }
  },

  getById: async (id) => {
    try {
      const response = await api.get(`/things/${id}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch device' 
      };
    }
  },

  update: async (id, thingData) => {
    try {
      const response = await api.put(`/things/${id}`, thingData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to update device' 
      };
    }
  },

  delete: async (id) => {
    if (isDemoMode()) {
      const things = getDemoData('things');
      const filteredThings = things.filter(thing => thing.id !== id);
      setDemoData('things', filteredThings);
      return { success: true };
    }
    
    try {
      await api.delete(`/things/${id}`);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to delete device' 
      };
    }
  }
};

// Channels API calls
export const channels = {
  getAll: async (offset = 0, limit = 20) => {
    if (isDemoMode()) {
      const channels = getDemoData('channels');
      const paginatedChannels = channels.slice(offset, offset + limit);
      return { 
        success: true, 
        data: { 
          channels: paginatedChannels, 
          total: channels.length,
          offset,
          limit 
        } 
      };
    }
    
    try {
      const response = await api.get(`/channels?offset=${offset}&limit=${limit}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch channels' 
      };
    }
  },

  create: async (channelData) => {
    if (isDemoMode()) {
      const channels = getDemoData('channels');
      const newChannel = {
        id: generateId(),
        ...channelData,
        created_at: new Date().toISOString(),
        status: channelData.status || 'enabled'
      };
      channels.push(newChannel);
      setDemoData('channels', channels);
      return { success: true, data: newChannel };
    }
    
    try {
      const response = await api.post('/channels', channelData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to create channel' 
      };
    }
  },

  getById: async (id) => {
    try {
      const response = await api.get(`/channels/${id}`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch channel' 
      };
    }
  },

  update: async (id, channelData) => {
    try {
      const response = await api.put(`/channels/${id}`, channelData);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to update channel' 
      };
    }
  },

  delete: async (id) => {
    if (isDemoMode()) {
      const channels = getDemoData('channels');
      const filteredChannels = channels.filter(channel => channel.id !== id);
      setDemoData('channels', filteredChannels);
      return { success: true };
    }
    
    try {
      await api.delete(`/channels/${id}`);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to delete channel' 
      };
    }
  },

  attachThing: async (channelId, thingId) => {
    try {
      await api.post(`/channels/${channelId}/things/${thingId}`);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to attach device to channel' 
      };
    }
  },

  detachThing: async (channelId, thingId) => {
    try {
      await api.delete(`/channels/${channelId}/things/${thingId}`);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to detach device from channel' 
      };
    }
  },

  getThings: async (channelId) => {
    try {
      const response = await api.get(`/channels/${channelId}/things`);
      return { success: true, data: response.data };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to fetch channel devices' 
      };
    }
  }
};

// Health check API calls
export const health = {
  checkService: async (serviceName, port = 9000) => {
    try {
      const serviceUrl = `http://localhost:${port}`;
      const response = await axios.get(`${serviceUrl}/health`, { timeout: 5000 });
      return { 
        success: true, 
        status: 'healthy', 
        service: serviceName,
        response: response.data 
      };
    } catch (error) {
      return { 
        success: false, 
        status: 'unhealthy', 
        service: serviceName,
        error: error.message 
      };
    }
  },

  checkAllServices: async () => {
    const services = [
      { name: 'Users Service', port: 9002 },
      { name: 'Things Service', port: 9000 },
      { name: 'HTTP Adapter', port: 8008 },
      { name: 'Auth Service', port: 9002 }
    ];

    const results = await Promise.allSettled(
      services.map(service => health.checkService(service.name, service.port))
    );

    return results.map((result, index) => ({
      service: services[index].name,
      ...(result.status === 'fulfilled' ? result.value : { success: false, status: 'error' })
    }));
  }
};

export default api;