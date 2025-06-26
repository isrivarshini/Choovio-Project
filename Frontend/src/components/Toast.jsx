import React, { useState, useEffect } from 'react';
import { CheckCircle, XCircle, AlertCircle, Info } from 'lucide-react';

const Toast = ({ type = 'info', message, isVisible, onClose, duration = 5000 }) => {
  const [isShowing, setIsShowing] = useState(false);

  useEffect(() => {
    if (isVisible) {
      setIsShowing(true);
      const timer = setTimeout(() => {
        setIsShowing(false);
        setTimeout(onClose, 300); // Wait for animation to complete
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [isVisible, duration, onClose]);

  const typeConfig = {
    success: {
      icon: CheckCircle,
      bgColor: 'bg-green-50',
      borderColor: 'border-green-200',
      textColor: 'text-green-800',
      iconColor: 'text-green-500'
    },
    error: {
      icon: XCircle,
      bgColor: 'bg-red-50',
      borderColor: 'border-red-200',
      textColor: 'text-red-800',
      iconColor: 'text-red-500'
    },
    warning: {
      icon: AlertCircle,
      bgColor: 'bg-yellow-50',
      borderColor: 'border-yellow-200',
      textColor: 'text-yellow-800',
      iconColor: 'text-yellow-500'
    },
    info: {
      icon: Info,
      bgColor: 'bg-blue-50',
      borderColor: 'border-blue-200',
      textColor: 'text-blue-800',
      iconColor: 'text-blue-500'
    }
  };

  const config = typeConfig[type];
  const Icon = config.icon;

  if (!isVisible) return null;

  return (
    <div className="fixed top-4 right-4 z-50">
      <div
        className={`
          flex items-center p-4 mb-4 border rounded-lg shadow-lg transition-all duration-300 transform
          ${config.bgColor} ${config.borderColor} ${config.textColor}
          ${isShowing ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'}
        `}
      >
        <Icon className={`w-5 h-5 mr-3 ${config.iconColor}`} />
        <p className="text-sm font-medium">{message}</p>
      </div>
    </div>
  );
};

// Toast Context and Hook
export const useToast = () => {
  const [toast, setToast] = useState({ isVisible: false, type: 'info', message: '' });

  const showToast = (type, message) => {
    setToast({ isVisible: true, type, message });
  };

  const hideToast = () => {
    setToast(prev => ({ ...prev, isVisible: false }));
  };

  return {
    toast,
    showToast,
    hideToast,
    showSuccess: (message) => showToast('success', message),
    showError: (message) => showToast('error', message),
    showWarning: (message) => showToast('warning', message),
    showInfo: (message) => showToast('info', message)
  };
};

export default Toast;