//go:build dnp3_ffi

/*
 * opendnp3_c.cpp — implementation of the flat C API declared in opendnp3_c.h,
 * on top of the opendnp3 (Apache 2.0) C++ master stack.
 *
 * Built by cgo as part of the dnp3 package under `-tags dnp3_ffi`. The
 * //go:build constraint above keeps it out of the default (stub) build, which
 * has no C compiler step.
 *
 * Standalone compile-check (no cgo needed):
 *   g++ -std=c++17 -fPIC \
 *       -I third_party/opendnp3/x86_64-unknown-linux-gnu/include \
 *       -c dnp3/opendnp3_c.cpp -o /tmp/opendnp3_c.o
 */

#include "opendnp3_c.h"

#include <chrono>
#include <cstring>
#include <memory>
#include <new>
#include <string>
#include <thread>
#include <vector>

#include "opendnp3/DNP3Manager.h"
#include "opendnp3/channel/ChannelRetry.h"
#include "opendnp3/channel/IChannel.h"
#include "opendnp3/channel/IChannelListener.h"
#include "opendnp3/channel/IPEndpoint.h"
#include "opendnp3/app/ClassField.h"
#include "opendnp3/app/GroupVariationID.h"
#include "opendnp3/app/MeasurementTypes.h"
#include "opendnp3/app/OctetString.h"
#include "opendnp3/app/parsing/ICollection.h"
#include "opendnp3/gen/ChannelState.h"
#include "opendnp3/logging/ILogHandler.h"
#include "opendnp3/logging/LogLevels.h"
#include "opendnp3/master/HeaderInfo.h"
#include "opendnp3/master/IMaster.h"
#include "opendnp3/master/IMasterApplication.h"
#include "opendnp3/master/IMasterScan.h"
#include "opendnp3/master/ISOEHandler.h"
#include "opendnp3/master/MasterStackConfig.h"
#include "opendnp3/master/TaskConfig.h"
#include "opendnp3/util/TimeDuration.h"
#include "opendnp3/util/UTCTimestamp.h"

using namespace opendnp3;

namespace
{

uint64_t now_ms()
{
    using namespace std::chrono;
    return static_cast<uint64_t>(
        duration_cast<milliseconds>(system_clock::now().time_since_epoch()).count());
}

// Map a single opendnp3 LogLevel flag to our coarse severity (0=err..3=debug).
int severity_of(const LogLevel& level)
{
    if (level.value & flags::ERR.value)
        return 0;
    if (level.value & flags::WARN.value)
        return 1;
    if (level.value & flags::INFO.value)
        return 2;
    return 3;
}

// --- ILogHandler -----------------------------------------------------------

class ShimLogHandler final : public ILogHandler
{
public:
    ShimLogHandler(const odc_callbacks* cbs, void* log_ctx) : cbs(cbs), log_ctx(log_ctx) {}

    void log(ModuleId, const char* id, LogLevel level, char const* location,
             char const* message) override
    {
        (void)id;
        (void)location;
        if (cbs && cbs->on_log)
            cbs->on_log(log_ctx, severity_of(level), message ? message : "");
    }

private:
    const odc_callbacks* cbs;
    void* log_ctx;
};

// --- IChannelListener ------------------------------------------------------

class ShimChannelListener final : public IChannelListener
{
public:
    ShimChannelListener(const odc_callbacks* cbs, void* ctx) : cbs(cbs), ctx(ctx) {}

    void OnStateChange(ChannelState state) override
    {
        if (cbs && cbs->on_channel_state)
            cbs->on_channel_state(ctx, static_cast<int>(state));
    }

private:
    const odc_callbacks* cbs;
    void* ctx;
};

// --- IMasterApplication ----------------------------------------------------
//
// All callbacks except IUTCTimeSource::Now() have usable defaults. We don't
// drive time sync (timeSyncMode = None), but Now() is pure virtual so it must
// be implemented; return wall-clock so it's sane if ever consulted.

class ShimMasterApplication final : public IMasterApplication
{
public:
    UTCTimestamp Now() override
    {
        return UTCTimestamp(now_ms());
    }
};

// --- ISOEHandler -----------------------------------------------------------

class ShimSOEHandler final : public ISOEHandler
{
public:
    ShimSOEHandler(const odc_callbacks* cbs, void* ctx) : cbs(cbs), ctx(ctx) {}

    void BeginFragment(const ResponseInfo&) override {}
    void EndFragment(const ResponseInfo&) override {}

    void Process(const HeaderInfo& info, const ICollection<Indexed<Binary>>& values) override
    {
        if (!cbs || !cbs->on_binary)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<Binary>& it) {
            cbs->on_binary(ctx, it.index, it.value.value ? 1 : 0, it.value.flags.value,
                           it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<DoubleBitBinary>>& values) override
    {
        if (!cbs || !cbs->on_double_bit)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<DoubleBitBinary>& it) {
            cbs->on_double_bit(ctx, it.index, static_cast<int>(it.value.value), it.value.flags.value,
                               it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<BinaryOutputStatus>>& values) override
    {
        if (!cbs || !cbs->on_binary_output_status)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<BinaryOutputStatus>& it) {
            cbs->on_binary_output_status(ctx, it.index, it.value.value ? 1 : 0, it.value.flags.value,
                                         it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<Counter>>& values) override
    {
        if (!cbs || !cbs->on_counter)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<Counter>& it) {
            cbs->on_counter(ctx, it.index, it.value.value, it.value.flags.value,
                            it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<FrozenCounter>>& values) override
    {
        if (!cbs || !cbs->on_frozen_counter)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<FrozenCounter>& it) {
            cbs->on_frozen_counter(ctx, it.index, it.value.value, it.value.flags.value,
                                   it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<Analog>>& values) override
    {
        if (!cbs || !cbs->on_analog)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<Analog>& it) {
            cbs->on_analog(ctx, it.index, it.value.value, it.value.flags.value,
                           it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<AnalogOutputStatus>>& values) override
    {
        if (!cbs || !cbs->on_analog_output_status)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<AnalogOutputStatus>& it) {
            cbs->on_analog_output_status(ctx, it.index, it.value.value, it.value.flags.value,
                                         it.value.time.value, static_cast<int>(it.value.time.quality), rt);
        });
    }

    void Process(const HeaderInfo& info, const ICollection<Indexed<OctetString>>& values) override
    {
        if (!cbs || !cbs->on_octet_string)
            return;
        const int rt = info.isEventVariation ? 1 : 0;
        values.ForeachItem([this, rt](const Indexed<OctetString>& it) {
            const Buffer b = it.value.ToBuffer();
            cbs->on_octet_string(ctx, it.index, b.data, b.length, rt);
        });
    }

    // Variations we don't surface to the gateway: drained but ignored.
    void Process(const HeaderInfo&, const ICollection<Indexed<TimeAndInterval>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<BinaryCommandEvent>>&) override {}
    void Process(const HeaderInfo&, const ICollection<Indexed<AnalogCommandEvent>>&) override {}
    void Process(const HeaderInfo&, const ICollection<DNPTime>&) override {}

private:
    const odc_callbacks* cbs;
    void* ctx;
};

} // namespace

// --- opaque structs --------------------------------------------------------

struct odc_manager
{
    std::unique_ptr<DNP3Manager> manager;
    std::shared_ptr<ShimLogHandler> logHandler;
    odc_callbacks cbs;
    int32_t logLevels;
};

struct odc_master
{
    std::shared_ptr<IChannel> channel;
    std::shared_ptr<IMaster> master;
    std::shared_ptr<ShimSOEHandler> soe;
    std::shared_ptr<ShimChannelListener> listener;
    std::shared_ptr<ShimMasterApplication> app;
    std::vector<std::shared_ptr<IMasterScan>> scans;
};

// --- C API -----------------------------------------------------------------

extern "C" {

odc_manager* odc_manager_create(uint32_t concurrency, odc_callbacks cbs, void* log_ctx,
                                int32_t log_level_mask)
{
    auto* m = new (std::nothrow) odc_manager();
    if (!m)
        return nullptr;
    m->cbs = cbs;
    m->logLevels = log_level_mask ? log_level_mask : levels::NORMAL.get_value();

    uint32_t threads = concurrency ? concurrency : std::thread::hardware_concurrency();
    if (threads == 0)
        threads = 1;

    try {
        m->logHandler = std::make_shared<ShimLogHandler>(&m->cbs, log_ctx);
        m->manager = std::make_unique<DNP3Manager>(threads, m->logHandler);
    } catch (...) {
        delete m;
        return nullptr;
    }
    return m;
}

void odc_manager_destroy(odc_manager* mgr)
{
    if (!mgr)
        return;
    if (mgr->manager)
        mgr->manager->Shutdown();
    delete mgr;
}

odc_master* odc_manager_add_master(odc_manager* mgr, const char* id, odc_master_config cfg,
                                   void* ctx)
{
    if (!mgr || !mgr->manager)
        return nullptr;

    auto* mst = new (std::nothrow) odc_master();
    if (!mst)
        return nullptr;

    const std::string sid = id ? id : "master";

    try {
        mst->soe = std::make_shared<ShimSOEHandler>(&mgr->cbs, ctx);
        mst->listener = std::make_shared<ShimChannelListener>(&mgr->cbs, ctx);
        mst->app = std::make_shared<ShimMasterApplication>();

        std::vector<IPEndpoint> hosts;
        hosts.push_back(IPEndpoint(cfg.host ? cfg.host : "127.0.0.1", cfg.port));

        mst->channel = mgr->manager->AddTCPClient(sid, LogLevels(mgr->logLevels),
                                                  ChannelRetry::Default(), hosts, "0.0.0.0",
                                                  mst->listener);
        if (!mst->channel) {
            delete mst;
            return nullptr;
        }

        MasterStackConfig stack;
        stack.master.responseTimeout =
            TimeDuration::Milliseconds(cfg.response_timeout_ms ? cfg.response_timeout_ms : 5000);
        stack.master.disableUnsolOnStartup = cfg.disable_unsol_on_startup != 0;
        stack.master.unsolClassMask = ClassField(false, cfg.unsol_class1 != 0,
                                                 cfg.unsol_class2 != 0, cfg.unsol_class3 != 0);
        stack.master.startupIntegrityClassMask =
            ClassField(cfg.startup_integrity_class0 != 0, cfg.startup_integrity_class1 != 0,
                       cfg.startup_integrity_class2 != 0, cfg.startup_integrity_class3 != 0);
        stack.link.LocalAddr = cfg.master_address;
        stack.link.RemoteAddr = cfg.outstation_address;
        stack.link.KeepAliveTimeout =
            TimeDuration::Milliseconds(cfg.keep_alive_ms ? cfg.keep_alive_ms : 60000);

        mst->master = mst->channel->AddMaster(sid, mst->soe, mst->app, stack);
        if (!mst->master) {
            mst->channel->Shutdown();
            delete mst;
            return nullptr;
        }
    } catch (...) {
        if (mst->channel)
            mst->channel->Shutdown();
        delete mst;
        return nullptr;
    }
    return mst;
}

int odc_master_add_class_scan(odc_master* mst, int class0, int class1, int class2, int class3,
                              uint32_t period_ms)
{
    if (!mst || !mst->master)
        return 1;
    try {
        auto scan = mst->master->AddClassScan(
            ClassField(class0 != 0, class1 != 0, class2 != 0, class3 != 0),
            TimeDuration::Milliseconds(period_ms), mst->soe);
        if (!scan)
            return 1;
        mst->scans.push_back(scan);
    } catch (...) {
        return 1;
    }
    return 0;
}

int odc_master_add_all_objects_scan(odc_master* mst, uint8_t group, uint8_t variation,
                                    uint32_t period_ms)
{
    if (!mst || !mst->master)
        return 1;
    try {
        auto scan = mst->master->AddAllObjectsScan(GroupVariationID(group, variation),
                                                   TimeDuration::Milliseconds(period_ms), mst->soe);
        if (!scan)
            return 1;
        mst->scans.push_back(scan);
    } catch (...) {
        return 1;
    }
    return 0;
}

int odc_master_scan_classes(odc_master* mst, int class0, int class1, int class2, int class3)
{
    if (!mst || !mst->master)
        return 1;
    try {
        mst->master->ScanClasses(ClassField(class0 != 0, class1 != 0, class2 != 0, class3 != 0),
                                 mst->soe);
    } catch (...) {
        return 1;
    }
    return 0;
}

int odc_master_enable(odc_master* mst)
{
    if (!mst || !mst->master)
        return 1;
    return mst->master->Enable() ? 0 : 1;
}

int odc_master_disable(odc_master* mst)
{
    if (!mst || !mst->master)
        return 1;
    return mst->master->Disable() ? 0 : 1;
}

void odc_master_destroy(odc_master* mst)
{
    if (!mst)
        return;
    // Shutting down the channel tears down its bound master too.
    if (mst->channel)
        mst->channel->Shutdown();
    delete mst;
}

} // extern "C"
