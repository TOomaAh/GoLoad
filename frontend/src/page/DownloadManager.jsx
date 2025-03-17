import React, { useState, useEffect } from 'react';
import Header from '../components/Header';
import Tabs from '../components/Tabs';
import DownloadList from '../components/DownloadList';
import Footer from '../components/Footer';
import NewDownloadModal from '../components/NewDownloadModal';
import SettingsPage from '../components/Settings';
import { useDownloads } from '../hooks/useDownloads';
import { useSettings } from '../hooks/useSettings';

const DownloadManager = () => {
  const [activeTab, setActiveTab] = useState('tous');
  const [showNewDownloadModal, setShowNewDownloadModal] = useState(false);
  const [showSettingsPage, setShowSettingsPage] = useState(false);
  
  const { 
    downloads, 
    isLoading, 
    apiError, 
    fetchDownloads, 
    handleAddDownload, 
    handleDownloadAction, 
    handleRemoveDownload 
  } = useDownloads(activeTab);
  
  const { 
    settings, 
    darkMode, 
    handleUpdateSettings 
  } = useSettings();

  // Recharger les téléchargements lors du changement d'onglet
  useEffect(() => {
    fetchDownloads();
  }, [activeTab, fetchDownloads]);

  // Si la page de paramètres est active, afficher celle-ci plutôt que la page principale
  if (showSettingsPage) {
    return (
      <SettingsPage 
        settings={settings} 
        darkMode={darkMode} 
        onUpdateSettings={handleUpdateSettings} 
        onBackClick={() => setShowSettingsPage(false)} 
      />
    );
  }

  return (
    <div className={`flex flex-col h-screen ${darkMode ? 'bg-gray-900 text-gray-100' : 'bg-gray-50 text-gray-800'}`}>
      <Header 
        darkMode={darkMode} 
        onNewDownloadClick={() => setShowNewDownloadModal(true)} 
        onSettingsClick={() => setShowSettingsPage(true)} 
      />
      
      {/* Affichage des erreurs API */}
      {apiError && (
        <div className={`${darkMode ? 'bg-red-900 border-red-700 text-red-200' : 'bg-red-100 border-red-500 text-red-700'} border-l-4 p-4 m-4`} role="alert">
          <p>{apiError}</p>
        </div>
      )}
      
      <Tabs 
        activeTab={activeTab} 
        onTabChange={setActiveTab} 
        darkMode={darkMode} 
      />
      
      <DownloadList 
        downloads={downloads} 
        isLoading={isLoading} 
        darkMode={darkMode} 
        onDownloadAction={handleDownloadAction} 
        onRemoveDownload={handleRemoveDownload} 
      />
      
      <Footer 
        downloads={downloads} 
        darkMode={darkMode} 
        onRefresh={fetchDownloads} 
      />
      
      {showNewDownloadModal && (
        <NewDownloadModal 
          darkMode={darkMode} 
          onClose={() => setShowNewDownloadModal(false)} 
          onAddDownload={handleAddDownload} 
        />
      )}
    </div>
  );
};

export default DownloadManager;