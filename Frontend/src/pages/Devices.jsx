import React, { useState, useEffect } from 'react';
import { Plus, Smartphone, Key, Eye, EyeOff, Copy, Wifi, Battery, Search } from 'lucide-react';
import { things } from '../api/api';
import Modal from '../components/Modal';
import LoadingSpinner from '../components/LoadingSpinner';
import { useToast } from '../components/Toast';
import Toast from '../components/Toast';

const Devices = () => {
  const [devicesList, setDevicesList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isKeyModalOpen, setIsKeyModalOpen] = useState(false);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [showKey, setShowKey] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    metadata: {}
  });
  const [formLoading, setFormLoading] = useState(false);
  const { toast, showSuccess, showError, hideToast } = useToast();

  useEffect(() => {
    fetchDevices();
  }, []);

  const fetchDevices = async () => {
    setLoading(true);
    try {
      const result = await things.getAll(0, 50);
      if (result.success) {
        setDevicesList(result.data.things || []);
      } else {
        showError(result.error || 'Failed to fetch devices');
      }
    } catch (error) {
      showError('An error occurred while fetching devices');
    } finally {
      setLoading(false);
    }
  };

  const handleAddDevice = () => {
    setFormData({
      name: '',
      metadata: {}
    });
    setIsModalOpen(true);
  };

  const handleViewKey = (device) => {
    setSelectedDevice(device);
    setShowKey(false);
    setIsKeyModalOpen(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormLoading(true);

    try {
      const result = await things.create(formData);
      if (result.success) {
        showSuccess('Device created successfully');
        setIsModalOpen(false);
        setSelectedDevice(result.data);
        setIsKeyModalOpen(true);
        fetchDevices();
      } else {
        showError(result.error || 'Failed to create device');
      }
    } catch (error) {
      showError('An error occurred while creating device');
    } finally {
      setFormLoading(false);
    }
  };

  const handleInputChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showSuccess('Copied to clipboard');
  };

  const filteredDevices = devicesList.filter(device =>
    device.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    device.id?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getDeviceStatus = () => {
    // Simulate device status
    const statuses = ['online', 'offline', 'maintenance'];
    return statuses[Math.floor(Math.random() * statuses.length)];
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'online':
        return 'bg-green-100 text-green-800';
      case 'offline':
        return 'bg-red-100 text-red-800';
      case 'maintenance':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      <Toast {...toast} onClose={hideToast} />
      
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Devices Management</h1>
          <p className="text-gray-600 mt-1">Register and manage IoT devices</p>
        </div>
        <button
          onClick={handleAddDevice}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          <span>Add Device</span>
        </button>
      </div>

      {/* Search and Stats */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex-1 relative max-w-md">
            <Search className="w-5 h-5 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search devices..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div className="flex items-center space-x-6">
            <div className="text-center">
              <p className="text-2xl font-bold text-gray-900">{devicesList.length}</p>
              <p className="text-sm text-gray-600">Total Devices</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-green-600">
                {Math.floor(devicesList.length * 0.8)}
              </p>
              <p className="text-sm text-gray-600">Online</p>
            </div>
          </div>
        </div>
      </div>

      {/* Devices Grid */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        {loading ? (
          <div className="p-12">
            <LoadingSpinner size="large" text="Loading devices..." />
          </div>
        ) : filteredDevices.length === 0 ? (
          <div className="p-12 text-center">
            <Smartphone className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No devices found</h3>
            <p className="text-gray-600 mb-4">
              {searchTerm ? 'No devices match your search criteria.' : 'Get started by adding your first device.'}
            </p>
            {!searchTerm && (
              <button
                onClick={handleAddDevice}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Add Device
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 p-6">
            {filteredDevices.map((device) => {
              const status = getDeviceStatus();
              return (
                <div key={device.id} className="border border-gray-200 rounded-xl p-6 hover:shadow-lg transition-shadow">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                        <Smartphone className="w-6 h-6 text-blue-600" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-gray-900">{device.name || 'Unnamed Device'}</h3>
                        <p className="text-sm text-gray-500">ID: {device.id}</p>
                      </div>
                    </div>
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(status)}`}>
                      {status}
                    </span>
                  </div>

                  <div className="space-y-3 mb-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Wifi className="w-4 h-4 text-gray-400" />
                        <span className="text-sm text-gray-600">Connection</span>
                      </div>
                      <span className="text-sm font-medium text-gray-900">
                        {status === 'online' ? 'Connected' : 'Disconnected'}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Battery className="w-4 h-4 text-gray-400" />
                        <span className="text-sm text-gray-600">Battery</span>
                      </div>
                      <span className="text-sm font-medium text-gray-900">
                        {Math.floor(Math.random() * 100)}%
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                    <span className="text-xs text-gray-500">
                      Created: {device.created_at ? new Date(device.created_at).toLocaleDateString() : 'N/A'}
                    </span>
                    <button
                      onClick={() => handleViewKey(device)}
                      className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
                    >
                      <Key className="w-4 h-4" />
                      <span>View Key</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Add Device Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Add New Device"
        size="medium"
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
              Device Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Enter device name"
            />
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start space-x-3">
              <Smartphone className="w-5 h-5 text-blue-600 mt-0.5" />
              <div>
                <h4 className="text-sm font-medium text-blue-900">Device Registration</h4>
                <p className="text-sm text-blue-700 mt-1">
                  After creating the device, you'll receive access credentials that should be configured on your IoT device.
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={() => setIsModalOpen(false)}
              className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={formLoading}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center space-x-2"
            >
              {formLoading && <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>}
              <span>Create Device</span>
            </button>
          </div>
        </form>
      </Modal>

      {/* Device Key Modal */}
      <Modal
        isOpen={isKeyModalOpen}
        onClose={() => setIsKeyModalOpen(false)}
        title="Device Access Key"
        size="large"
      >
        {selectedDevice && (
          <div className="space-y-6">
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <div className="flex items-start space-x-3">
                <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                  <Key className="w-4 h-4 text-green-600" />
                </div>
                <div>
                  <h4 className="text-sm font-medium text-green-900">Device Created Successfully</h4>
                  <p className="text-sm text-green-700 mt-1">
                    Your device has been registered. Use the credentials below to configure your IoT device.
                  </p>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Device ID</label>
                <div className="flex items-center space-x-2">
                  <input
                    type="text"
                    value={selectedDevice.id}
                    readOnly
                    className="flex-1 px-3 py-2 bg-gray-50 border border-gray-300 rounded-lg"
                  />
                  <button
                    onClick={() => copyToClipboard(selectedDevice.id)}
                    className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Access Key</label>
                <div className="flex items-center space-x-2">
                  <input
                    type={showKey ? 'text' : 'password'}
                    value={selectedDevice.key || selectedDevice.credentials?.secret || 'key-' + selectedDevice.id}
                    readOnly
                    className="flex-1 px-3 py-2 bg-gray-50 border border-gray-300 rounded-lg font-mono text-sm"
                  />
                  <button
                    onClick={() => setShowKey(!showKey)}
                    className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    {showKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                  <button
                    onClick={() => copyToClipboard(selectedDevice.key || selectedDevice.credentials?.secret || 'key-' + selectedDevice.id)}
                    className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-amber-900 mb-2">Security Notice</h4>
              <ul className="text-sm text-amber-800 space-y-1">
                <li>• Store these credentials securely</li>
                <li>• Do not share the access key publicly</li>
                <li>• The key cannot be recovered if lost</li>
              </ul>
            </div>

            <div className="flex items-center justify-end pt-4">
              <button
                onClick={() => setIsKeyModalOpen(false)}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Devices;