#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>

class QNetworkAccessManager;
class QNetworkReply;
class QMediaPlayer;
class QAudioOutput;
class QListWidget;

QT_BEGIN_NAMESPACE
namespace Ui {
class MainWindow;
}
QT_END_NAMESPACE

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private slots:
    void cargarRolas(const QString &busqueda = "");
    void procesaRes(QNetworkReply *respuesta);
    void minarDirectorio();
    void agregaPersona();
    void consultaPersona(int id);
    void agregaGrupo();
    void cargarGrupos();
    void consultaGrupo(int id);
    void cargarIntegrantes(int idGrupo, QListWidget *lista);

private:
    Ui::MainWindow *ui;
    QNetworkAccessManager *administradorRed;
    QMediaPlayer *reproductor;
    QAudioOutput *salidaAudio;
    QString      rutaActual;
    QString      rutaReproduciendo;
};
#endif // MAINWINDOW_H
