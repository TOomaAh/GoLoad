import { useState, useEffect, useCallback } from 'react';
import { 
  GetAllDownloads, 
  GetDownloadsByStatus, 
  PauseDownload, 
  ResumeDownload,
  CancelDownload, 
  AddDownload, 
  RemoveDownload 
} from '../../wailsjs/go/app/App';

export const useDownloads = (activeTab) => {
  const [downloads, setDownloads] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [apiError, setApiError] = useState(null);
  const [newDownloadUrl, setNewDownloadUrl] = useState('');

  // Charger les téléchargements
  const fetchDownloads = useCallback(async () => {
    try {
      setIsLoading(true);
      let data;

      if (activeTab === 'tous') {
        try {
          data = await GetAllDownloads();
        } catch (error) {
          console.error('Erreur lors du chargement de tous les téléchargements:', error);
          data = [];
        }
      } else {
        // Convertir le nom d'onglet en statut
        const statusMap = {
          'actifs': 'downloading',
          'terminés': 'completed',
          'queued': 'queued',
          'pause': 'paused'
        };
        data = await GetDownloadsByStatus(statusMap[activeTab] || '');
      }
      
      setDownloads(data);
      setApiError(null);
    } catch (error) {
      console.error('Erreur lors du chargement des téléchargements:', error);
      setApiError('Impossible de charger les téléchargements');
    } finally {
      setIsLoading(false);
    }
  }, [activeTab]);

  // Fonction pour ajouter un nouveau téléchargement
  const handleAddDownload = async (url) => {
    if (!url.trim()) return;
    
    try {
      await AddDownload(url, '', '');
      
      // Actualiser la liste des téléchargements
      fetchDownloads();
      setNewDownloadUrl('');
      setApiError(null);
      return true;
    } catch (error) {
      console.error('Erreur lors de l\'ajout du téléchargement:', error);
      setApiError(`Impossible d'ajouter le téléchargement: ${error.message || error}`);
      return false;
    }
  };

  // Gérer les actions sur les téléchargements
  const handleDownloadAction = async (id, action) => {
    try {
      switch (action) {
        case 'pause':
          await PauseDownload(id);
          break;
        case 'resume':
          await ResumeDownload(id);
          break;
        case 'cancel':
          await CancelDownload(id);
          break;
        default:
          throw new Error(`Action non reconnue: ${action}`);
      }
      
      // Actualiser la liste des téléchargements
      fetchDownloads();
      setApiError(null);
    } catch (error) {
      console.error(`Erreur lors de l'action ${action}:`, error);
      setApiError(`Impossible d'effectuer l'action ${action}: ${error.message || error}`);
    }
  };

  // Supprimer un téléchargement
  const handleRemoveDownload = async (id, deleteFile = false) => {
    try {
      await RemoveDownload(id, deleteFile);
      
      // Actualiser la liste des téléchargements
      fetchDownloads();
      setApiError(null);
    } catch (error) {
      console.error('Erreur lors de la suppression du téléchargement:', error);
      setApiError(`Impossible de supprimer le téléchargement: ${error.message || error}`);
    }
  };

  // Charger les données au démarrage
  useEffect(() => {
    fetchDownloads();
    
    // Configuration de l'écouteur pour les mises à jour en temps réel
    if (window.runtime) {
      window.runtime.EventsOn('download:status', () => {
        fetchDownloads(); // Rafraîchir les téléchargements à chaque mise à jour
      });
      
      return () => {
        window.runtime.EventsOff('download:status');
      };
    }
  }, [fetchDownloads]);

  return {
    downloads,
    isLoading,
    apiError,
    newDownloadUrl,
    setNewDownloadUrl,
    fetchDownloads,
    handleAddDownload,
    handleDownloadAction,
    handleRemoveDownload
  };
};