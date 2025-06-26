import React, { useState, useEffect } from 'react';
import { Plus, Radio, Link, Unlink, Search, Smartphone } from 'lucide-react';
import { channels, things } from '../api/api';
import Modal from '../components/Modal';
import LoadingSpinner from '../components/LoadingSpinner';
import { useToast } from '../components/Toast';
import Toast from '../components/Toast';

const Channels = () => {
  const [channelsList, setChannelsList] = useState([]);
  const [devicesList, setDevicesList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isLinkModalOpen, setIsLinkModalOpen] = useState(false);
  const [selectedChannel, setSelectedChannel] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    metadata: {}
  });
  const [formLoading, setFormLoading] = useState(false);
  const [selectedDevices, setSelectedDevices] = useState([]);
  const { toast, showSuccess, showError, hideToast } = useToast();

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      // Fetch channels and devices in parallel
      const [channelsResult, devicesResult] = await Promise.all([
        channels.getAll(0, 50),
        things.getAll(0, 50)
      ]);

      if (channelsResult.success) {
        setChannelsList(channelsResult.data.channels || []);
      } else {
        showError(channelsResult.error || 'Failed to fetch channels');
      }

      if (devicesResult.success) {
        setDevicesList(devicesResult.data.things || []);
      } else {
        showError(devicesResult.error || 'Failed to fetch devices');
      }
    } catch (error) {
      showError('An error occurred while fetching data');
    } finally {
      setLoading(false);
    }
  };

  const handleAddChannel = () => {
    setFormData({
      name: '',
      description: '',
      metadata: {}
    });
    setIsModalOpen(true);
  };

  const handleLinkDevices = (channel) => {
    setSelectedChannel(channel);
    setSelectedDevices([]);
    setIsLinkModalOpen(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormLoading(true);

    try {
      const result = await channels.create(formData);
      if (result.success) {
        showSuccess('Channel created successfully');
        setIsModalOpen(false);
        fetchData();
      } else {
        showError(result.error || 'Failed to create channel');
      }
    } catch (error) {
      showError('An error occurred while creating channel');
    } finally {
      setFormLoading(false);
    }
  };

  const handleLinkDevicesSubmit = async () => {
    if (!selectedChannel || selectedDevices.length === 0) {
      showError('Please select devices to link');
      return;
    }

    setFormLoading(true);
    try {
      const promises = selectedDevices.map(deviceId =>
        channels.attachThing(selectedChannel.id, deviceId)
      );

      const results = await Promise.all(promises);
      const successful = results.filter(r => r.success).length;
      const failed = results.length - successful;

      if (successful > 0) {
        showSuccess(`Successfully linked ${successful} device(s) to channel`);
      }
      if (failed > 0) {
        showError(`Failed to link ${failed} device(s)`);
      }

      setIsLinkModalOpen(false);
      fetchData();
    } catch (error) {
      showError('An error occurred while linking devices');
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

  const handleDeviceSelection = (deviceId, isSelected) => {
    if (isSelected) {
      setSelectedDevices([...selectedDevices, deviceId]);
    } else {
      setSelectedDevices(selectedDevices.filter(id => id !== deviceId));
    }
  };

  const filteredChannels = channelsList.filter(channel =>
    channel.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    channel.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getChannelStatus = () => {
    // Simulate channel status
    const statuses = ['active', 'inactive', 'error'];
    return statuses[Math.floor(Math.random() * statuses.length)];
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-gray-100 text-gray-800';
      case 'error':
        return 'bg-red-100 text-red-800';
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
          <h1 className="text-2xl font-bold text-gray-900">Channels Management</h1>
          <p className="text-gray-600 mt-1">Manage data channels and device connections</p>
        </div>
        <button
          onClick={handleAddChannel}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          <span>Add Channel</span>
        </button>
      </div>

      {/* Search and Stats */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex-1 relative max-w-md">
            <Search className="w-5 h-5 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search channels..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div className="flex items-center space-x-6">
            <div className="text-center">
              <p className="text-2xl font-bold text-gray-900">{channelsList.length}</p>
              <p className="text-sm text-gray-600">Total Channels</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-blue-600">
                {Math.floor(channelsList.length * 0.9)}
              </p>
              <p className="text-sm text-gray-600">Active</p>
            </div>
          </div>
        </div>
      </div>

      {/* Channels Grid */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        {loading ? (
          <div className="p-12">
            <LoadingSpinner size="large" text="Loading channels..." />
          </div>
        ) : filteredChannels.length === 0 ? (
          <div className="p-12 text-center">
            <Radio className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No channels found</h3>
            <p className="text-gray-600 mb-4">
              {searchTerm ? 'No channels match your search criteria.' : 'Get started by creating your first channel.'}
            </p>
            {!searchTerm && (
              <button
                onClick={handleAddChannel}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Add Channel
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 p-6">
            {filteredChannels.map((channel) => {
              const status = getChannelStatus();
              const connectedDevices = Math.floor(Math.random() * 5);
              
              return (
                <div key={channel.id} className="border border-gray-200 rounded-xl p-6 hover:shadow-lg transition-shadow">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                        <Radio className="w-6 h-6 text-purple-600" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-gray-900">{channel.name || 'Unnamed Channel'}</h3>
                        <p className="text-sm text-gray-500">ID: {channel.id}</p>
                      </div>
                    </div>
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(status)}`}>
                      {status}
                    </span>
                  </div>

                  {channel.description && (
                    <p className="text-sm text-gray-600 mb-4">{channel.description}</p>
                  )}

                  <div className="space-y-3 mb-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Smartphone className="w-4 h-4 text-gray-400" />
                        <span className="text-sm text-gray-600">Connected Devices</span>
                      </div>
                      <span className="text-sm font-medium text-gray-900">{connectedDevices}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Radio className="w-4 h-4 text-gray-400" />
                        <span className="text-sm text-gray-600">Messages Today</span>
                      </div>
                      <span className="text-sm font-medium text-gray-900">
                        {Math.floor(Math.random() * 1000)}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                    <span className="text-xs text-gray-500">
                      Created: {channel.created_at ? new Date(channel.created_at).toLocaleDateString() : 'N/A'}
                    </span>
                    <button
                      onClick={() => handleLinkDevices(channel)}
                      className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
                    >
                      <Link className="w-4 h-4" />
                      <span>Link Devices</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Add Channel Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Add New Channel"
        size="medium"
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
              Channel Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Enter channel name"
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
              Description (Optional)
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Enter channel description"
            />
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start space-x-3">
              <Radio className="w-5 h-5 text-blue-600 mt-0.5" />
              <div>
                <h4 className="text-sm font-medium text-blue-900">Channel Purpose</h4>
                <p className="text-sm text-blue-700 mt-1">
                  Channels act as communication pipelines between your devices and applications. You can link multiple devices to a single channel.
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
              <span>Create Channel</span>
            </button>
          </div>
        </form>
      </Modal>

      {/* Link Devices Modal */}
      <Modal
        isOpen={isLinkModalOpen}
        onClose={() => setIsLinkModalOpen(false)}
        title={`Link Devices to ${selectedChannel?.name}`}
        size="large"
      >
        <div className="space-y-6">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start space-x-3">
              <Link className="w-5 h-5 text-blue-600 mt-0.5" />
              <div>
                <h4 className="text-sm font-medium text-blue-900">Device Linking</h4>
                <p className="text-sm text-blue-700 mt-1">
                  Select devices to connect to this channel. Connected devices can send data through this channel.
                </p>
              </div>
            </div>
          </div>

          <div className="max-h-64 overflow-y-auto border border-gray-200 rounded-lg">
            {devicesList.length === 0 ? (
              <div className="p-8 text-center">
                <Smartphone className="w-8 h-8 text-gray-300 mx-auto mb-2" />
                <p className="text-gray-600">No devices available to link</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-200">
                {devicesList.map((device) => (
                  <div key={device.id} className="flex items-center space-x-3 p-4 hover:bg-gray-50">
                    <input
                      type="checkbox"
                      id={`device-${device.id}`}
                      checked={selectedDevices.includes(device.id)}
                      onChange={(e) => handleDeviceSelection(device.id, e.target.checked)}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                    />
                    <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                      <Smartphone className="w-4 h-4 text-blue-600" />
                    </div>
                    <div className="flex-1">
                      <label htmlFor={`device-${device.id}`} className="text-sm font-medium text-gray-900 cursor-pointer">
                        {device.name || 'Unnamed Device'}
                      </label>
                      <p className="text-xs text-gray-500">ID: {device.id}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="flex items-center justify-between pt-4">
            <p className="text-sm text-gray-600">
              {selectedDevices.length} device(s) selected
            </p>
            <div className="flex items-center space-x-3">
              <button
                onClick={() => setIsLinkModalOpen(false)}
                className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleLinkDevicesSubmit}
                disabled={formLoading || selectedDevices.length === 0}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center space-x-2"
              >
                {formLoading && <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>}
                <span>Link Devices</span>
              </button>
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Channels;