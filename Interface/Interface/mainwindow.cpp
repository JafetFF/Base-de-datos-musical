#include "mainwindow.h"
#include "./ui_mainwindow.h"
#include <QNetworkAccessManager>
#include <QNetworkRequest>
#include <QNetworkReply>
#include <QUrl>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QDebug>
#include <QTableWidgetItem>
#include <QUrlQuery>
#include <QFileDialog>
#include <QMessageBox>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , administradorRed(new QNetworkAccessManager(this))
{
    ui->setupUi(this);
    connect(administradorRed, &QNetworkAccessManager::finished, this, &MainWindow::procesaRes);
    cargarRolas();
    connect(ui->songsButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->songsPage);
    });
    connect(ui->searchLineEdit, &QLineEdit::textChanged, this, &MainWindow::cargarRolas);
    connect(ui->foldersButton, &QPushButton::clicked, this, &MainWindow::minarDirectorio);
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::cargarRolas(const QString &busqueda)
{
    QUrl
        direccion("http://127.0.0.1:8080/api/rolas");
    QUrlQuery parametros;
    parametros.addQueryItem("q", busqueda);
    direccion.setQuery(parametros);
    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}

void MainWindow::procesaRes(QNetworkReply *respuesta)
{
    QByteArray datos = respuesta->readAll();

    if (respuesta->url().path() == "/api/minar") {
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::critical(
                this,
                "Ocurrió un error al minar",
                QString::fromUtf8(datos)
                );
        } else {
            QJsonDocument documento = QJsonDocument::fromJson(datos);

            if (documento.isObject()) {
                QJsonObject resultado = documento.object();

                QString estado = resultado["status"].toString();
                QString mensaje = resultado["message"].toString();

                if (estado == "ok") {
                    QMessageBox::information(
                        this,
                        "Minado completado",
                        mensaje
                        );

                    cargarRolas(ui->searchLineEdit->text());
                } else {
                    QMessageBox::warning(
                        this,
                        "Resultado del minado",
                        mensaje
                        );
                }
            }
        }

        respuesta->deleteLater();
        return;
    }

    if (respuesta->error() != QNetworkReply::NoError) {
        qDebug() << "Error al consultar la API:"
                 << respuesta->errorString();

        respuesta->deleteLater();
        return;
    }

    QJsonDocument documento = QJsonDocument::fromJson(datos);

    if (!documento.isArray()) {
        qDebug() << "La respuesta no contiene un arreglo json";

        respuesta->deleteLater();
        return;
    }

    QJsonArray rolas = documento.array();

    ui->songsTable->setRowCount(0);

    for (const QJsonValue &valor : rolas) {
        QJsonObject rola = valor.toObject();

        int fila = ui->songsTable->rowCount();

        ui->songsTable->insertRow(fila);
        ui->songsTable->setItem(fila, 0, new QTableWidgetItem(rola["title"].toString()));
        ui->songsTable->setItem(fila, 1, new QTableWidgetItem(rola["artist"].toString()));
        ui->songsTable->setItem(fila, 2, new QTableWidgetItem(rola["album"].toString()));
        ui->songsTable->setItem(fila, 3, new QTableWidgetItem(QString::number(rola["year"].toInt())));
    }

    qDebug() << "Rolas recibidas:" << rolas.size();

    respuesta->deleteLater();
}

void MainWindow::minarDirectorio()
{
    QString directorio = QFileDialog::getExistingDirectory(
        this,
        "Seleccionar carpeta de rolas"
        );

    if (directorio.isEmpty()) {
        return;
    }

    QUrl direccion("http://127.0.0.1:8080/api/minar");
    QNetworkRequest peticion(direccion);

    peticion.setHeader(
        QNetworkRequest::ContentTypeHeader,
        "application/json"
        );

    QJsonObject solicitud;
    solicitud["directorio"] = directorio;

    QJsonDocument documento(solicitud);
    QByteArray datos = documento.toJson();

    administradorRed->post(peticion, datos);
}
