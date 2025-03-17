import React, { useState } from 'react';

const NewDownloadModal = ({ darkMode, onClose, onAddDownload }) => {
  const [newDownloadUrl, setNewDownloadUrl] = useState('');

  const handleSubmit = async () => {
    const success = await onAddDownload(newDownloadUrl);
    if (success) {
      setNewDownloadUrl('');
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className={`${darkMode ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow-lg max-w-md w-full p-6`}>
        <h2 className="text-xl font-bold mb-4">Nouveau téléchargement</h2>
        
        <div className="mb-4">
          <label className="block text-sm font-medium mb-2" htmlFor="downloadUrl">
            URL du fichier
          </label>
          <input
            type="text"
            id="downloadUrl"
            className={`w-full px-3 py-2 border ${
              darkMode 
                ? 'border-gray-600 bg-gray-700 focus:border-blue-500' 
                : 'border-gray-300 focus:border-blue-500'
            } rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500`}
            placeholder="https://exemple.com/fichier.zip"
            value={newDownloadUrl}
            onChange={(e) => setNewDownloadUrl(e.target.value)}
          />
        </div>
        
        <div className="flex justify-end space-x-3">
          <button
            onClick={onClose}
            className={`px-4 py-2 border ${
              darkMode 
                ? 'border-gray-600 hover:bg-gray-700' 
                : 'border-gray-300 hover:bg-gray-100'
            } rounded-md transition-colors`}
          >
            Annuler
          </button>
          <button
            onClick={handleSubmit}
            className={`px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors ${
              !newDownloadUrl.trim() ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            disabled={!newDownloadUrl.trim()}
          >
            Télécharger
          </button>
        </div>
      </div>
    </div>
  );
};

export default NewDownloadModal;