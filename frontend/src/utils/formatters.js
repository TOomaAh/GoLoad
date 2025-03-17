/**
 * Formate la vitesse de téléchargement en unités lisibles
 * @param {number} bytesPerSecond - Vitesse en octets par seconde
 * @returns {string} - Vitesse formatée (ex: "1.5 MB/s")
 */
export const formatSpeed = (bytesPerSecond) => {
    if (bytesPerSecond === 0) return '0 B/s';
    
    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    const i = Math.floor(Math.log(bytesPerSecond) / Math.log(1024));
    return `${(bytesPerSecond / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
  };
  
  /**
   * Retourne le libellé correspondant au statut d'un téléchargement
   * @param {string} status - Statut du téléchargement
   * @returns {string} - Libellé du statut
   */
  export const getStatusLabel = (status) => {
    switch (status) {
      case 'completed': return 'Terminé';
      case 'downloading': return 'En téléchargement';
      case 'paused': return 'En pause';
      case 'queued': return 'En attente';
      case 'cancelled': return 'Annulé';
      case 'error': return 'Erreur';
      default: return status;
    }
  };
  
  /**
   * Renvoie les classes CSS pour un statut donné
   * @param {string} status - Statut du téléchargement
   * @param {boolean} darkMode - Si le mode sombre est activé
   * @returns {string} - Classes CSS
   */
  export const getStatusClasses = (status, darkMode) => {
    switch (status) {
      case 'completed':
        return darkMode ? 'bg-green-900 text-green-300' : 'bg-green-100 text-green-800';
      case 'downloading':
        return darkMode ? 'bg-blue-900 text-blue-300' : 'bg-blue-100 text-blue-800';
      case 'paused':
        return darkMode ? 'bg-yellow-900 text-yellow-300' : 'bg-yellow-100 text-yellow-800';
      case 'cancelled':
      case 'error':
        return darkMode ? 'bg-red-900 text-red-300' : 'bg-red-100 text-red-800';
      default:
        return darkMode ? 'bg-gray-700 text-gray-300' : 'bg-gray-100 text-gray-800';
    }
  };
  
  /**
   * Renvoie la couleur de la barre de progression en fonction du statut
   * @param {string} status - Statut du téléchargement
   * @returns {string} - Classe CSS pour la couleur
   */
  export const getProgressBarColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-500';
      case 'paused': return 'bg-yellow-500';
      case 'error': return 'bg-red-500';
      default: return 'bg-blue-500';
    }
  };