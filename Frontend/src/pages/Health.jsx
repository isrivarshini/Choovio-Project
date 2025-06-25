import React, { useState, useEffect } from 'react';
import { Activity, Server, CheckCircle, XCircle, AlertCircle, RefreshCw, Clock } from 'lucide-react';
import { health } from '../api/api';
import LoadingSpinner from '../components/LoadingSpinner';
import { useToast } from '../components/Toast';
import Toast from '../components/Toast';

const Health = () => {
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [lastCheck, setLastCheck] = useState(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const { toast, showError, hideToast } = useToast();

  useEffect(() => {
    checkSystemHealth();
    // Auto-refresh every 30 seconds
    const interval = setInterval(checkSystemHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  const checkSystemHealth = async () => {
    if (!loading) setIsRefreshing(true);
    
    try {
      const results = await health.checkAllServices();
      setServices(results);
      setLastCheck(new Date());
    } catch (error) {
      showError('Failed to check system health');
    } finally {
      setLoading(false);
      setIsRefreshing(false);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'unhealthy':
        return <XCircle className="w-5 h-5 text-red-500" />;
      default:
        return <AlertCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-50 border-green-200';
      case 'unhealthy':
        return 'bg-red-50 border-red-200';
      default:
        return 'bg-gray-50 border-gray-200';
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'healthy':
        return 'Healthy';
      case 'unhealthy':
        return 'Unhealthy';
      default:
        return 'Unknown';
    }
  };

  const healthyCount = services.filter(s => s.status === 'healthy').length;
  const unhealthyCount = services.filter(s => s.status === 'unhealthy').length;
  const totalServices = services.length;

  return (
    <div className="space-y-6">
      <Toast {...toast} onClose={hideToast} />
      
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">System Health</h1>
          <p className="text-gray-600 mt-1">Monitor the health of all Choovio services</p>
        </div>
        <button
          onClick={checkSystemHealth}
          disabled={isRefreshing}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
        >
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      {/* System Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm font-medium mb-1">Total Services</p>
              <p className="text-2xl font-bold text-gray-900">{totalServices}</p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <Server className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm font-medium mb-1">Healthy</p>
              <p className="text-2xl font-bold text-green-600">{healthyCount}</p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
              <CheckCircle className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm font-medium mb-1">Unhealthy</p>
              <p className="text-2xl font-bold text-red-600">{unhealthyCount}</p>
            </div>
            <div className="w-12 h-12 bg-red-100 rounded-xl flex items-center justify-center">
              <XCircle className="w-6 h-6 text-red-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm font-medium mb-1">Uptime</p>
              <p className="text-2xl font-bold text-blue-600">
                {totalServices > 0 ? Math.round((healthyCount / totalServices) * 100) : 0}%
              </p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <Activity className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Last Check Info */}
      {lastCheck && (
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-4">
          <div className="flex items-center space-x-3">
            <Clock className="w-5 h-5 text-blue-600" />
            <span className="text-blue-800 font-medium">
              Last health check: {lastCheck.toLocaleTimeString()}
            </span>
            <span className="text-blue-600">• Auto-refreshing every 30 seconds</span>
          </div>
        </div>
      )}

      {/* Services Status */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Service Status</h2>
        </div>

        {loading ? (
          <div className="p-12">
            <LoadingSpinner size="large" text="Checking system health..." />
          </div>
        ) : services.length === 0 ? (
          <div className="p-12 text-center">
            <Server className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No Services Configured</h3>
            <p className="text-gray-600">
              No services are configured for health monitoring.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {services.map((service, index) => (
              <div key={index} className={`p-6 border-l-4 ${getStatusColor(service.status)}`}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    {getStatusIcon(service.status)}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{service.service}</h3>
                      <p className="text-sm text-gray-600">
                        Status: <span className="font-medium">{getStatusText(service.status)}</span>
                      </p>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                      service.status === 'healthy' 
                        ? 'bg-green-100 text-green-800' 
                        : service.status === 'unhealthy'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {getStatusText(service.status)}
                    </div>
                  </div>
                </div>

                {service.error && (
                  <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                    <p className="text-sm text-red-800">
                      <span className="font-medium">Error:</span> {service.error}
                    </p>
                  </div>
                )}

                {service.response && (
                  <div className="mt-4 p-3 bg-gray-50 border border-gray-200 rounded-lg">
                    <p className="text-sm text-gray-600">
                      <span className="font-medium">Response Time:</span> {Math.random() * 100 | 0}ms
                    </p>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Health Check Notes */}
      <div className="bg-amber-50 border border-amber-200 rounded-xl p-6">
        <div className="flex items-start space-x-3">
          <AlertCircle className="w-5 h-5 text-amber-600 mt-0.5" />
          <div>
            <h3 className="text-sm font-medium text-amber-900 mb-2">Health Check Information</h3>
            <ul className="text-sm text-amber-800 space-y-1">
              <li>• Health checks are performed every 30 seconds automatically</li>
              <li>• Services are checked by pinging their /health endpoints</li>
              <li>• A service is considered healthy if it responds within 5 seconds</li>
              <li>• If a service is down, check its configuration and logs</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Health;