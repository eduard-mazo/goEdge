/*
 * outstation_sim.cpp — non-interactive DNP3 outstation for smoke-testing the
 * goMqttDnp3 master binding. Links against the vendored opendnp3 (Apache 2.0)
 * static lib. Listens on a TCP port, serves a small database, and mutates a
 * few points on a timer so the master sees both static (integrity) data and
 * change events.
 *
 * Link addressing matches the gateway defaults used by the smoke harness:
 *   outstation LocalAddr = 1024, RemoteAddr (master) = 1.
 *
 * Build (see scripts/dnp3-smoke.sh):
 *   g++ -std=c++17 -I third_party/opendnp3/<triple>/include \
 *       scripts/sim/outstation_sim.cpp \
 *       third_party/opendnp3/<triple>/lib/libopendnp3.a \
 *       -lssl -lcrypto -lpthread -o /tmp/outstation_sim
 *
 * Usage: outstation_sim [port]   (default 20000). Runs until killed.
 */
#include <opendnp3/ConsoleLogger.h>
#include <opendnp3/DNP3Manager.h>
#include <opendnp3/channel/PrintingChannelListener.h>
#include <opendnp3/logging/LogLevels.h>
#include <opendnp3/outstation/DefaultOutstationApplication.h>
#include <opendnp3/outstation/SimpleCommandHandler.h>
#include <opendnp3/outstation/UpdateBuilder.h>

#include <chrono>
#include <cstdlib>
#include <iostream>
#include <string>
#include <thread>

using namespace opendnp3;

int main(int argc, char* argv[])
{
    const uint16_t port = (argc > 1) ? static_cast<uint16_t>(std::atoi(argv[1])) : 20000;

    // NORMAL is enough; uncomment ALL_COMMS to see frames.
    const auto logLevels = levels::NORMAL;

    DNP3Manager manager(1, ConsoleLogger::Create());

    std::shared_ptr<IChannel> channel;
    try {
        channel = manager.AddTCPServer("sim", logLevels, ServerAcceptMode::CloseExisting,
                                       IPEndpoint("0.0.0.0", port), PrintingChannelListener::Create());
    } catch (const std::exception& e) {
        std::cerr << "outstation_sim: failed to bind :" << port << " — " << e.what() << std::endl;
        return 1;
    }

    // 5 of each type; default clazz = Class1 (events generated on change).
    DatabaseConfig db(5);

    OutstationStackConfig config(db);
    config.outstation.eventBufferConfig = EventBufferConfig::AllTypes(100);
    config.outstation.params.allowUnsolicited = true;
    config.link.LocalAddr = 1024; // this outstation's address
    config.link.RemoteAddr = 1;   // the master's address
    config.link.KeepAliveTimeout = TimeDuration::Max();

    auto app = DefaultOutstationApplication::Create();
    auto outstation = channel->AddOutstation("sim-ostn", SuccessCommandHandler::Create(), app, config);

    // Seed initial static values so an integrity poll returns meaningful data.
    {
        UpdateBuilder b;
        for (uint16_t i = 0; i < 5; ++i) {
            b.Update(Binary(i % 2 == 0, Flags(0x01), app->Now()), i);
            b.Update(Analog(10.0 + i, Flags(0x01), app->Now()), i);
            b.Update(Counter(100 + i, Flags(0x01), app->Now()), i);
            b.Update(DoubleBitBinary(DoubleBit::DETERMINED_ON, Flags(0x01), app->Now()), i);
        }
        outstation->Apply(b.Build());
    }

    outstation->Enable();
    std::cerr << "outstation_sim: listening on 0.0.0.0:" << port
              << " (outstation addr 1024, master addr 1)" << std::endl;

    // Mutate a handful of points on a timer to drive change events.
    bool binary = false;
    double analog = 10.0;
    uint32_t count = 100;
    DoubleBit dbit = DoubleBit::DETERMINED_ON;
    uint8_t octet = 1;

    while (true) {
        std::this_thread::sleep_for(std::chrono::milliseconds(1500));
        UpdateBuilder b;
        binary = !binary;
        analog += 1.0;
        ++count;
        dbit = (dbit == DoubleBit::DETERMINED_ON) ? DoubleBit::DETERMINED_OFF : DoubleBit::DETERMINED_ON;
        b.Update(Binary(binary, Flags(0x01), app->Now()), 0);
        b.Update(Analog(analog, Flags(0x01), app->Now()), 0);
        b.Update(Analog(analog * 2, Flags(0x01), app->Now()), 1);
        b.Update(Counter(count, Flags(0x01), app->Now()), 0);
        b.Update(DoubleBitBinary(dbit, Flags(0x01), app->Now()), 0);
        OctetString os(Buffer(&octet, 1));
        b.Update(os, 0);
        ++octet;
        outstation->Apply(b.Build());
    }

    return 0;
}
