#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>

class QNetworkAccessManager;
class QNetworkReply;

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

private:
    Ui::MainWindow *ui;
    QNetworkAccessManager *administradorRed;
};
#endif // MAINWINDOW_H
