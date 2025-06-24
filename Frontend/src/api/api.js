import axios from 'axios';

// Base configuration for Magistrala API
const BASE_URL = ''; // Default Magistrala API endpoint

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
      // Always try real API call
      const response = await api.post('/tokens', { email, password });
      const token = response.data.access_token;
      authToken = token;
      localStorage.setItem('magistrala_token', token);
      return { success: true, token };
    } catch (error) {
      console.error('Login error:', error);
      // Provide detailed error information
      let errorMessage = 'Login failed';
      if (error.code === 'ECONNREFUSED' || error.code === 'ERR_NETWORK') {
        errorMessage = 'Cannot connect to Magistrala server. Please check if the server is running on localhost:9000';
      } else if (error.response?.status === 401) {
        errorMessage = 'Invalid email or password';
      } else if (error.response?.status === 404) {
        errorMessage = 'Magistrala API endpoint not found. Please check server configuration.';
      } else if (error.response?.data?.message) {
        errorMessage = error.response.data.message;
      } else if (error.message) {
        errorMessage = error.message;
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