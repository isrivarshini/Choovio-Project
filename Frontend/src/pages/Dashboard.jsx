import React, { useState, useEffect } from 'react';
import { 
  Users, 
  Smartphone, 
  Radio, 
  Activity, 
  TrendingUp, 
  TrendingDown,
  Wifi,
  Battery,
  RefreshCw
} from 'lucide-react';
import { users, things, channels, health } from '../api/api';
import LoadingSpinner from '../components/LoadingSpinner';

const Dashboard = () => {
  const [stats, setStats] = useState({
    users: { count: 0, loading: true },
    devices: { count: 0, loading: true },
    channels: { count: 0, loading: true },
    health: { services: [], loading: true }
  });

  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchStats = async () => {
    setIsRefreshing(true);

    // Fetch users count
    try {
      const usersResult = await users.getAll(0, 1);
      if (usersResult.success) {
        setStats(prev => ({
          ...prev,
          users: { count: usersResult.data.total || 0, loading: false }
        }));
      }
    } catch (error) {
      setStats(prev => ({
        ...prev,
        users: { count: 0, loading: false }
      }));
    }

    // Fetch devices count
    try {
      const devicesResult = await things.getAll(0, 1);
      if (devicesResult.success) {
        setStats(prev => ({
          ...prev,
          devices: { count: devicesResult.data.total || 0, loading: false }
        }));
      }
    } catch (error) {
      setStats(prev => ({
        ...prev,
        devices: { count: 0, loading: false }
      }));
    }

    // Fetch channels count
    try {
      const channelsResult = await channels.getAll(0, 1);
      if (channelsResult.success) {
        setStats(prev => ({
          ...prev,
          channels: { count: channelsResult.data.total || 0, loading: false }
        }));
      }
    } catch (error) {
      setStats(prev => ({
        ...prev,
        channels: { count: 0, loading: false }
      }));
    }

    // Fetch health status
    try {
      const healthResults = await health.checkAllServices();
      setStats(prev => ({
        ...prev,
        health: { services: healthResults, loading: false }
      }));
    } catch (error) {
      setStats(prev => ({
        ...prev,
        health: { services: [], loading: false }
      }));
    }

    setIsRefreshing(false);
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const StatCard = ({ title, value, icon: Icon, loading, trend, trendValue, color = 'blue' }) => {
    const colorClasses = {
      blue: 'from-blue-500 to-blue-600',
      green: 'from-green-500 to-green-600',
      purple: 'from-purple-500 to-purple-600',
      orange: 'from-orange-500 to-orange-600'
    };

    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-gray-600 text-sm font-medium mb-1">{title}</p>
            {loading ? (
              <div className="h-8 w-16 bg-gray-200 rounded animate-pulse"></div>
            ) : (
              <p className="text-2xl font-bold text-gray-900">{value.toLocaleString()}</p>
            )}
            {trend && trendValue && (
              <div className={`flex items-center mt-2 text-sm ${trend === 'up' ? 'text-green-600' : 'text-red-600'}`}>
                {trend === 'up' ? <TrendingUp className="w-4 h-4 mr-1" /> : <TrendingDown className="w-4 h-4 mr-1" />}
                <span>{trendValue}% from last month</span>
              </div>
            )}
          </div>
          <div className={`w-12 h-12 bg-gradient-to-r ${colorClasses[color]} rounded-xl flex items-center justify-center`}>
            <Icon className="w-6 h-6 text-white" />
          </div>
        </div>
      </div>
    );
  };

  const ServiceStatus = ({ service, status, loading }) => {
    const statusColors = {
      healthy: 'bg-green-100 text-green-800 border-green-200',
      unhealthy: 'bg-red-100 text-red-800 border-red-200',
      error: 'bg-gray-100 text-gray-800 border-gray-200'
    };

    const statusIcons = {
      healthy: '🟢',
      unhealthy: '🔴',
      error: '⚪'
    };

    return (
      <div className="flex items-center justify-between py-3 px-4 bg-gray-50 rounded-lg">
        <div className="flex items-center space-x-3">
          <span className="text-lg">{statusIcons[status] || '⚪'}</span>
          <span className="font-medium text-gray-900">{service}</span>
        </div>
        {loading ? (
          <div className="w-16 h-6 bg-gray-200 rounded animate-pulse"></div>
        ) : (
          <span className={`px-3 py-1 rounded-full text-xs font-medium border ${statusColors[status] || statusColors.error}`}>
            {status || 'Unknown'}
          </span>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-600 mt-1">Welcome to your Magistrala admin dashboard</p>
        </div>
        <button
          onClick={fetchStats}
          disabled={isRefreshing}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
        >
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Real-time Status Banner */}
      <div className="bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200 rounded-xl p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
            <span className="font-medium text-amber-800">Real-time Device Monitoring</span>
            <span className="text-amber-600">• Updated Live</span>
          </div>
          <span className="text-sm text-amber-700">0 messages received</span>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value={stats.users.count}
          icon={Users}
          loading={stats.users.loading}
          trend="up"
          trendValue="12"
          color="blue"
        />
        <StatCard
          title="Active Devices"
          value={stats.devices.count}
          icon={Smartphone}
          loading={stats.devices.loading}
          trend="up"
          trendValue="8"
          color="green"
        />
        <StatCard
          title="Data Channels"
          value={stats.channels.count}
          icon={Radio}
          loading={stats.channels.loading}
          trend="up"
          trendValue="15"
          color="purple"
        />
        <StatCard
          title="System Health"
          value={stats.health.services.filter(s => s.status === 'healthy').length}
          icon={Activity}
          loading={stats.health.loading}
          color="orange"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* System Health */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">System Services</h2>
            <Activity className="w-5 h-5 text-gray-400" />
          </div>
          
          <div className="space-y-3">
            {stats.health.loading ? (
              <div className="space-y-3">
                {[1, 2, 3, 4].map(i => (
                  <div key={i} className="flex items-center justify-between py-3 px-4 bg-gray-50 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <div className="w-4 h-4 bg-gray-200 rounded-full animate-pulse"></div>
                      <div className="w-24 h-4 bg-gray-200 rounded animate-pulse"></div>
                    </div>
                    <div className="w-16 h-6 bg-gray-200 rounded animate-pulse"></div>
                  </div>
                ))}
              </div>
            ) : stats.health.services.length > 0 ? (
              stats.health.services.map((service, index) => (
                <ServiceStatus
                  key={index}
                  service={service.service}
                  status={service.status}
                  loading={false}
                />
              ))
            ) : (
              <div className="text-center py-8 text-gray-500">
                <Activity className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>No service data available</p>
              </div>
            )}
          </div>
        </div>

        {/* Quick Metrics */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">Live Metrics</h2>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm text-gray-600">Live</span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <Wifi className="w-8 h-8 text-green-500 mx-auto mb-2" />
              <p className="text-2xl font-bold text-gray-900">3</p>
              <p className="text-sm text-gray-600">Online Devices</p>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <Battery className="w-8 h-8 text-blue-500 mx-auto mb-2" />
              <p className="text-2xl font-bold text-gray-900">85%</p>
              <p className="text-sm text-gray-600">Avg Battery</p>
            </div>
          </div>

          <div className="mt-6 p-4 bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-700">Data Throughput</p>
                <p className="text-lg font-semibold text-gray-900">1.2 MB/min</p>
              </div>
              <TrendingUp className="w-8 h-8 text-green-500" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;